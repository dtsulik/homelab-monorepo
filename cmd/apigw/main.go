package main

import (
	"io"
	"net/http"
	_ "net/http/pprof"

	"gif-doggo/internal/helpers"
	"gif-doggo/internal/jaegerexport"
	"gif-doggo/internal/logger"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"golang.org/x/exp/slices"
)

var (
	tracer_name = "doggo-apigw"
	status_ep   = helpers.GetEnv("STATUS_ENDPOINT", "http://gif-doggo-status")
	http_client *http.Client
)

func main() {
	tc_ep := helpers.GetEnv("TRACECOLLECTOR_ENDPOINT", "http://jaeger:14268/api/traces")
	logger.Infow("Sending traces to " + tc_ep)

	tp, err := jaegerexport.JaegerTraceProvider(tracer_name, tc_ep)
	if err != nil {
		logger.Fatalw("Failed to create tracer provider", "error", err)
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tr := otelhttp.NewTransport(http.DefaultTransport, otelhttp.WithTracerProvider(tp))
	http_client = &http.Client{Transport: tr}

	logger.Infow("Starting server", "port", 80)
	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, request *http.Request) {})
	mux.HandleFunc("/livez", func(w http.ResponseWriter, request *http.Request) {})
	mux.Handle("/", otelhttp.NewHandler(
		apigw{}, "apigw-http-request",
		otelhttp.WithTracerProvider(tp)))

	err = http.ListenAndServe(":80", mux)
	if err != nil {
		logger.Fatalw("Failed to start server", "error", err)
	}
}

type apigw struct{}

type ep struct {
	methods []string
	path    string
	host    string
}

var endpoints = map[string]ep{
	"status": {
		methods: []string{"GET"},
		path:    "/status",
		host:    status_ep,
	},
}

func (apigw) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	logger.Infow("Handling api request.", "ep", request.URL.Path)

	ctx := request.Context()

	for _, ep := range endpoints {
		if slices.Contains(ep.methods, request.Method) && request.URL.Path == ep.path {
			req, _ := http.NewRequestWithContext(ctx, request.Method, request.URL.Path, nil)
			resp, err := http_client.Do(req)
			if err != nil {
				logger.Errorw("Failed to reach endpoint handler.",
					"path", request.URL.Path,
					"method", request.Method,
					"handler", ep.host,
					"error", err)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(resp.StatusCode)
			w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
			io.Copy(w, resp.Body)
			return
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
	w.WriteHeader(http.StatusNotImplemented)
}

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
	// otel.SetTextMapPropagator(propagation.TraceContext{})

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

func (apigw) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	logger.Infow("Handling api request.", "ep", request.URL.Path)

	ctx := request.Context()

	switch {
	case request.Method == "GET" && request.URL.Path == "/status":
		logger.Infow("Getting status", "url", status_ep)
		req, _ := http.NewRequestWithContext(ctx, "GET", status_ep, nil)
		resp, err := http_client.Do(req)

		defer resp.Body.Close()

		if err != nil {
			logger.Errorw("Failed to get status", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(resp.StatusCode)
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		io.Copy(w, resp.Body)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
}

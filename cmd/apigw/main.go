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
	"go.opentelemetry.io/otel/attribute"
)

var (
	tracer_name = "doggo-apigw"
	status_ep   = helpers.GetEnv("STATUS_ENDPOINT", "http://gif-doggo-status")
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
	http_client := &http.Client{Transport: tr}

	logger.Infow("Starting server", "port", 80)
	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, request *http.Request) {})
	mux.HandleFunc("/livez", func(w http.ResponseWriter, request *http.Request) {})
	mux.HandleFunc("/", func(w http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()

		ctx, span := otel.Tracer(tracer_name).Start(request.Context(), "apigw-request")
		defer span.End()

		span.SetAttributes(attribute.String("url", request.URL.String()))
		span.SetAttributes(attribute.String("method", request.Method))

		logger.Infow("Handling api request.", "ep", request.URL.Path)

		switch {
		case request.Method == "GET" && request.URL.Path == "/status":
			logger.Infow("Getting status", "url", status_ep)
			req, _ := http.NewRequestWithContext(ctx, "GET", status_ep, nil)
			// propagation.TraceContext{}.Inject(ctx, propagation.HeaderCarrier(req.Header))
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
	})

	err = http.ListenAndServe(":80", mux)
	if err != nil {
		logger.Fatalw("Failed to start server", "error", err)
	}
}

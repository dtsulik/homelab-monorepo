package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	_ "net/http/pprof"

	"gif-doggo/internal/helpers"
	"gif-doggo/internal/jaegerexport"
	"gif-doggo/internal/logger"

	"github.com/go-redis/redis/extra/redisotel/v9"
	"github.com/go-redis/redis/v9"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var redis_client *redis.Client
var tracer_name = "doggo-random"

func init() {
	redis_client = redis.NewClient(&redis.Options{
		Addr: helpers.GetEnv("REDIS_ENDPOINT", "redis:6379"),
	})
	if err := redisotel.InstrumentTracing(redis_client); err != nil {
		logger.Fatalw("Unable to start redis otel")
	}
}

func main() {
	tc_ep := helpers.GetEnv("TRACECOLLECTOR_ENDPOINT", "http://jaeger:14268/api/traces")
	logger.Infow("Sending traces to " + tc_ep)

	tp, err := jaegerexport.JaegerTraceProvider(tracer_name, tc_ep)
	if err != nil {
		logger.Fatalw("Failed to create tracer provider", "error", err)
	}

	otel.SetTracerProvider(tp)

	logger.Infow("Starting server", "port", 80)
	http.HandleFunc("/readyz", func(w http.ResponseWriter, request *http.Request) {})
	http.HandleFunc("/livez", func(w http.ResponseWriter, request *http.Request) {})
	http.HandleFunc("/", handle_root)
	err = http.ListenAndServe(":80", nil)
	if err != nil {
		logger.Fatalw("Failed to start server", "error", err)
	}
}

func handle_root(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	ctx, span := otel.Tracer(tracer_name).Start(context.TODO(), "random-file-request")
	defer span.End()

	logger.Infow("Received request", "method", request.Method, "url", request.URL)

	file, err := retrieve_file(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "image/gif")
	_, err = io.Copy(w, file)
	if err != nil {
		logger.Errorw("Unable to send file", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// TODO - replace with minio
func retrieve_file(ctx context.Context) (io.Reader, error) {
	_, span := otel.Tracer(tracer_name).Start(ctx, "random-file-retreival")
	defer span.End()

	rnd := redis_client.RandomKey(ctx).String()
	image_body, err := redis_client.Get(ctx, rnd).Bytes()
	if err != nil {
		logger.Errorw("Failed to retrieve file", "filename", rnd, "error", err)
		return nil, err
	}

	span.SetAttributes(attribute.String("request.filename", rnd))
	span.SetAttributes(attribute.Int("request.filesize", len(image_body)))

	return bytes.NewReader(image_body), nil
}

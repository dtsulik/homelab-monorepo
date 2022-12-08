package main

import (
	"context"
	"io"
	"net/http"
	_ "net/http/pprof"

	"gif-doggo/internal/jaegerexport"
	"gif-doggo/internal/logger"

	"github.com/go-redis/redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var redis_client *redis.Client
var tracer_name = "doggo-intake"

// TODO env vars here
func init() {
	redis_client = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
}

func main() {
	tp, err := jaegerexport.JaegerTraceProvider("http://jaeger:14268/api/traces")
	if err != nil {
		logger.Fatalw("Failed to create tracer provider", "error", err)
	}

	otel.SetTracerProvider(tp)

	logger.Infow("Starting server", "port", 80)
	http.HandleFunc("/readyz", func(w http.ResponseWriter, request *http.Request) {})
	http.HandleFunc("/livez", func(w http.ResponseWriter, request *http.Request) {})
	http.HandleFunc("/", handle_root)

	// TODO above cancel is useless if the server is blocking and not handling signals
	err = http.ListenAndServe(":80", nil)
	if err != nil {
		logger.Fatalw("Failed to start server", "error", err)
	}
}

func handle_root(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	ctx, span := otel.Tracer(tracer_name).Start(context.TODO(), "upload")
	defer span.End()

	logger.Infow("Received request", "method", request.Method, "url", request.URL)
	if _, ok := request.Header["Filename"]; !ok {
		logger.Errorw("Filename header not found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := receive_file(ctx, request.Body, "test")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// TODO - replace with minio
// FIXME - that 3600 is hardcoded still
func receive_file(ctx context.Context, body io.Reader, image_key string) error {
	_, span := otel.Tracer(tracer_name).Start(ctx, "file-upload")
	defer span.End()

	file, err := io.ReadAll(body)
	if err != nil {
		logger.Errorw("Failed to read body", "filename", image_key, "error", err)
		return err
	}

	span.SetAttributes(attribute.String("request.filename", image_key))
	span.SetAttributes(attribute.Int("request.filesize", len(file)))

	_, err = redis_client.Set(context.Background(), image_key, file, 3600).Result()
	if err != nil {
		logger.Errorw("Failed to save file", "filename", image_key, "error", err)
		return err
	}
	return nil
}

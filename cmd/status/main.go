package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"gif-doggo/internal/logger"

	"github.com/go-redis/redis/v9"
	"go.opentelemetry.io/otel"
)

var redis_client *redis.Client
var tracer_name = "doggo-requests"

func init() {
	redis_client = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
}

func main() {
	logger.Infow("Starting server", "port", 8080)
	http.HandleFunc("/", handle_root)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		logger.Fatalw("Failed to start server", "error", err)
	}
}

func handle_root(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	ctx, span := otel.Tracer(tracer_name).Start(context.Background(), "receive-requests")
	defer span.End()

	logger.Infow("Received request", "method", request.Method, "url", request.URL)
	uid := request.URL.Query().Get("uid")
	status := provide_status(ctx, uid)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "%s"}`, status)
}

func provide_status(ctx context.Context, uid string) string {
	_, span := otel.Tracer(tracer_name).Start(ctx, "retrieve-status")
	defer span.End()

	return redis_client.Get(context.Background(), uid).Val()
}

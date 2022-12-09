package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"time"

	"gif-doggo/internal/helpers"
	"gif-doggo/internal/jaegerexport"
	"gif-doggo/internal/logger"

	"github.com/go-redis/redis/extra/redisotel/v9"
	"github.com/go-redis/redis/v9"

	"github.com/google/uuid"

	"go.opentelemetry.io/otel"
)

var redis_client *redis.Client
var tracer_name = "doggo-requests"

// TODO env vars here
func init() {
	redis_client = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
	if err := redisotel.InstrumentTracing(redis_client); err != nil {
		logger.Fatalw("Unable to start redis otel")
	}
}

func main() {
	tc_ep := helpers.GetEnv("TRACECOLLECTOR_ENDPOINT", "http://jaeger:14268/api/traces")
	logger.Infow("Sending traces to " + tc_ep)

	tp, err := jaegerexport.JaegerTraceProvider(tc_ep)
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

type doggo_request struct {
	UUID       string   `json:"uuid"`
	Images     []string `json:"images"`
	Output     string   `json:"output"`
	Delays     []int    `json:"delays"`
	Expiration int      `json:"expiration"`
}

func handle_root(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	ctx, span := otel.Tracer(tracer_name).Start(context.Background(), "receive-requests")
	defer span.End()

	logger.Infow("Received request", "method", request.Method, "url", request.URL)

	body, err := io.ReadAll(request.Body)
	if err != nil {
		logger.Errorw("Invalid request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := doggo_request{}
	if err := json.Unmarshal(body, &req); err != nil {
		logger.Errorw("Invalid request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.UUID = uuid.New().String()
	err = publish_request(ctx, req)
	if err != nil {
		logger.Errorw("Failed to publish request", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": "%s"}`, req.UUID)
}

// TODO this is unhandled issue here, what happens if we publish the message but fail to update status? chicken and egg problem
func publish_request(ctx context.Context, req doggo_request) error {
	_, span := otel.Tracer(tracer_name).Start(ctx, "publish-request")
	defer span.End()

	// TODO add redis otel
	err := redis_client.Publish(ctx, "doggos", req).Err()
	if err != nil {
		logger.Errorw("Failed to submit request", "error", err)
		return err
	}

	err = redis_client.Set(ctx, req.UUID, "submitted", time.Duration(req.Expiration)).Err()
	if err != nil {
		logger.Errorw("Failed to update request status", "error", err)
		return err
	}
	return nil
}

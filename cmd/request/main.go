package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"

	"gif-doggo/internal/logger"

	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
)

var redis_client *redis.Client

// TODO env vars here
func init() {
	redis_client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
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

	logger.Infow("Received request", "method", request.Method, "url", request.URL)
	req_id := uuid.New()

	err := publish_request(req_id, request.Body)
	if err != nil {
		logger.Errorw("Failed to publish request", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": "%s"}`, req_id)
}

func publish_request(uid uuid.UUID, req_body io.ReadCloser) error {
	err := redis_client.Publish(context.Background(), "doggos", req_body).Err()
	if err != nil {
		logger.Errorw("Failed to submit request", "error", err)
		return err
	}

	// TODO set expiry from the request, infitinte doggos cost more money
	err = redis_client.Set(context.Background(), uid.String(), "submitted", 3600).Err()
	if err != nil {
		logger.Errorw("Failed to update request status", "error", err)
		return err
	}
	return nil
}

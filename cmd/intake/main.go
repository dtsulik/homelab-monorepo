package main

import (
	"context"
	"io"
	"net/http"
	_ "net/http/pprof"

	"gif-doggo/internal/logger"

	"github.com/go-redis/redis/v9"
)

var redis_client *redis.Client

// TODO env vars here
func init() {
	redis_client = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
}

func main() {
	logger.Infow("Starting server", "port", 80)
	http.HandleFunc("/", handle_root)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		logger.Fatalw("Failed to start server", "error", err)
	}
}

func handle_root(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	logger.Infow("Received request", "method", request.Method, "url", request.URL)
	if _, ok := request.Header["Filename"]; !ok {
		logger.Errorw("Filename header not found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := receive_file(request.Body, "test")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// TODO - replace with minio
// FIXME - that 3600 is hardcoded still
func receive_file(body io.Reader, image_key string) error {

	file, err := io.ReadAll(body)
	if err != nil {
		logger.Errorw("Failed to read body", "filename", image_key, "error", err)
		return err
	}
	_, err = redis_client.Set(context.Background(), image_key, file, 3600).Result()
	if err != nil {
		logger.Errorw("Failed to save file", "filename", image_key, "error", err)
		return err
	}
	return nil
}

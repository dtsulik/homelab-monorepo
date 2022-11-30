package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	_ "net/http/pprof"

	"gif-doggo/internal/logger"

	"github.com/go-redis/redis/v9"
)

var redis_client *redis.Client

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

	file, err := retrieve_file(request.Header.Get("Filename"))
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
func retrieve_file(image_key string) (io.Reader, error) {

	image_body, err := redis_client.Get(context.Background(), image_key).Bytes()
	if err != nil {
		logger.Errorw("Failed to retrieve file", "filename", image_key, "error", err)
		return nil, err
	}
	return bytes.NewReader(image_body), nil
}

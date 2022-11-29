package main

import (
	"fmt"
	"io"
	"net/http"

	"gif-doggo/internal/logger"

	"github.com/go-redis/redis/v9"
)

var redis_client *redis.Client

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
	b, _ := io.ReadAll(request.Body)
	fmt.Fprint(w, string(b))
}

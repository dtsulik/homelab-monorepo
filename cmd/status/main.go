package main

import (
	"context"
	"fmt"
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
	uid := request.URL.Query().Get("uid")
	status := provide_status(uid)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "%s"}`, status)
}

func provide_status(uid string) string {
	return redis_client.Get(context.Background(), uid).Val()
}

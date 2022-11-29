package main

import (
	"fmt"
	"io"
	"net/http"

	"gif-doggo/internal/logger"
)

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

package main

import (
	"fmt"
	"io"
	"net/http"

	"gif-doggo/pkg/logger"
)

func main() {
	logger.Log(logger.LogLevelInfo, "Starting server")
	http.HandleFunc("/", handle_root)
	http.ListenAndServe(":80", nil)
}

func handle_root(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	logger.Log(logger.LogLevelInfo, "Received request")
	b, _ := io.ReadAll(request.Body)
	fmt.Fprint(w, string(b))
}

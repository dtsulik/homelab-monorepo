package main

import (
	"io"
	"net/http"
	"os"

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
func receive_file(body io.ReadCloser, filename string) error {
	defer body.Close()

	file, err := os.Create(filename)
	if err != nil {
		logger.Errorw("Failed to create file", "filename", filename, "error", err)
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	if err != nil {
		logger.Errorw("Failed to write file", "error", err)
		return err
	}
	return nil
}

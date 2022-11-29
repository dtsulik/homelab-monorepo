package main

import (
	"io"
	"net/http"
	"os"

	"gif-doggo/internal/logger"
)

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
	defer file.Close()

	w.Header().Set("Content-Type", "image/gif")
	_, err = io.Copy(w, file)
	if err != nil {
		logger.Errorw("Unable to send file", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// TODO - replace with minio
func retrieve_file(filename string) (io.ReadCloser, error) {

	f, err := os.Open(filename)
	if err != nil {
		logger.Errorw("Failed to open file", "filename", filename, "error", err)
		return nil, err
	}
	return f, nil
}

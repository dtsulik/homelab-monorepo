package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	http.HandleFunc("/", handle_root)
	http.ListenAndServe(":80", nil)
}

func handle_root(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	b, _ := io.ReadAll(request.Body)
	fmt.Fprintf(w, string(b))
}

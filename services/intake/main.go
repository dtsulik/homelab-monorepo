package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	http.HandleFunc("/", upload)
	http.ListenAndServe(":80", nil)
}

func upload(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	b, _ := io.ReadAll(request.Body)
	fmt.Fprint(w, string(b))
}

package main

import (
	"net/http"

	"github.com/hacdias/filemanager"
	handlers "github.com/hacdias/filemanager/http"
)

var cfg *filemanager.Config

func handler(w http.ResponseWriter, r *http.Request) {
	handlers.ServeHTTP(w, r, cfg)
}

func main() {
	cfg = filemanager.New("D:\\TEST\\")

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

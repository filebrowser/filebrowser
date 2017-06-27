package main

import (
	"log"
	"net/http"

	"github.com/hacdias/filemanager"
)

var m *filemanager.FileManager

func handler(w http.ResponseWriter, r *http.Request) {
	_, err := m.ServeHTTP(w, r)
	if err != nil {
		log.Print(err)
	}
}

func main() {
	m = filemanager.New("D:\\TEST")
	m.SetBaseURL("/vaca")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

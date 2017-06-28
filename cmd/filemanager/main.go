package main

import (
	"log"
	"net/http"

	"github.com/hacdias/filemanager"
)

var m *filemanager.FileManager

func handler(w http.ResponseWriter, r *http.Request) {
	// TODO: review return codes and return 0 when everything works.

	code, err := m.ServeHTTP(w, r)
	if err != nil {
		log.Print(err)
	}

	if code != 0 {
		w.WriteHeader(code)
	}
}

func main() {
	m = filemanager.New("D:\\TEST")
	m.SetBaseURL("/vaca")
	m.Commands = []string{"git"}
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

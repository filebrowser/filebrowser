package filemanager

import (
	"errors"
	"net/http"
)

// Page is the base type for each page
type Page struct {
	GET, POST, PUT, DELETE         func(w http.ResponseWriter, r *http.Request) (int, error)
	DoGET, DoPOST, DoPUT, DoDELETE bool
}

// Route redirects the request for the respective method
func (p Page) Route(w http.ResponseWriter, r *http.Request) (int, error) {
	switch r.Method {
	case "DELETE":
		if p.DoDELETE {
			return p.DELETE(w, r)
		}
	case "POST":
		if p.DoPOST {
			return p.POST(w, r)
		}
	case "GET":
		if p.DoGET {
			return p.GET(w, r)
		}
	case "PUT":
		if p.DoPUT {
			return p.PUT(w, r)
		}
	}

	return http.StatusMethodNotAllowed, errors.New("Invalid method.")
}

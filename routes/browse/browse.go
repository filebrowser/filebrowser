package browse

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
)

var conf *config.Config

type response struct {
	Message  string `json:"message"`
	Location string `json:"location"`
}

// ServeHTTP is used to serve the content of Browse page using Browse middleware
// from Caddy. It handles the requests for DELETE, POST, GET and PUT related to
// /browse interface.
func ServeHTTP(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	conf = c
	// Removes the page main path from the URL
	r.URL.Path = strings.Replace(r.URL.Path, "/admin/browse", "", 1)

	switch r.Method {
	case "DELETE":
		return DELETE(w, r)
	case "POST":
		return POST(w, r)
	case "GET":
		return GET(w, r)
	case "PUT":
		return PUT(w, r)
	default:
		return http.StatusMethodNotAllowed, errors.New("Invalid method.")
	}
}

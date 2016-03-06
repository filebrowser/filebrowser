package browse

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
)

// ServeHTTP is used to serve the content of Browse page using Browse middleware
// from Caddy. It handles the requests for DELETE, POST, GET and PUT related to
// /browse interface.
func ServeHTTP(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Removes the page main path from the URL
	r.URL.Path = strings.Replace(r.URL.Path, "/admin/browse", "", 1)

	switch r.Method {
	case "DELETE":
		return DELETE(w, r, c)
	case "POST":
		return POST(w, r, c)
	case "GET":
		return GET(w, r, c)
	case "PUT":
		return PUT(w, r, c)
	default:
		return http.StatusMethodNotAllowed, errors.New("Invalid method.")
	}
}

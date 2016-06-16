package browse

import (
	"net/http"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
)

type response struct {
	Message  string `json:"message"`
	Location string `json:"location"`
}

// ServeHTTP is used to serve the content of Browse page using Browse middleware
// from Caddy. It handles the requests for DELETE, POST, GET and PUT related to
// /browse interface.
func ServeHTTP(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	r.URL.Path = strings.Replace(r.URL.Path, c.Admin+"/browse", "", 1)

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
		return http.StatusNotImplemented, nil
	}
}

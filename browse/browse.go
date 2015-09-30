package browse

import (
	"net/http"
	"strings"

	"github.com/hacdias/caddy-cms/config"
)

// ServeHTTP is used to serve the content of Browse page
// using Browse middleware from Caddy
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
	default:
		return 400, nil
	}
}

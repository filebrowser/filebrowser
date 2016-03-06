package editor

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
)

// ServeHTTP serves the editor page
func ServeHTTP(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	filename := strings.Replace(r.URL.Path, "/admin/edit/", "", 1)
	filename = c.Path + filename

	switch r.Method {
	case "POST":
		return POST(w, r, c, filename)
	case "GET":
		return GET(w, r, c, filename)
	default:
		return http.StatusMethodNotAllowed, errors.New("Invalid method.")
	}
}

package editor

import (
	"net/http"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
)

// ServeHTTP serves the editor page
func ServeHTTP(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	filename := strings.Replace(r.URL.Path, c.Admin+"/edit/", "", 1)
	filename = c.Path + filename

	switch r.Method {
	case http.MethodPost:
		return POST(w, r, c, filename)
	case http.MethodGet:
		return GET(w, r, c, filename)
	default:
		return http.StatusNotImplemented, nil
	}
}

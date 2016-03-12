package editor

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
)

var (
	filename string
	conf     *config.Config
)

// ServeHTTP serves the editor page
func ServeHTTP(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	conf = c
	filename = strings.Replace(r.URL.Path, "/admin/edit/", "", 1)
	filename = c.Path + filename

	switch r.Method {
	case "POST":
		return POST(w, r)
	case "GET":
		return GET(w, r)
	default:
		return http.StatusMethodNotAllowed, errors.New("Invalid method.")
	}
}

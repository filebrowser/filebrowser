package editor

import (
	"net/http"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
)

type editor struct {
	Name        string
	Class       string
	IsPost      bool
	Mode        string
	Content     string
	FrontMatter interface{}
	Config      *config.Config
}

// ServeHTTP serves the editor page
func ServeHTTP(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	filename := strings.Replace(r.URL.Path, "/admin/edit/", "", 1)

	switch r.Method {
	case "POST":
		return POST(w, r, c, filename)
	case "GET":
		return GET(w, r, c, filename)
	default:
		return 400, nil
	}
}

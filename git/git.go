package git

import (
	"errors"
	"net/http"

	"github.com/hacdias/caddy-hugo/config"
)

// ServeHTTP is used to serve the content of GIT API.
func ServeHTTP(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	switch r.Method {
	case "POST":
		return POST(w, r, c)
	default:
		return 400, errors.New("Invalid method.")
	}
}

package git

import (
	"errors"
	"net/http"

	"github.com/hacdias/caddy-hugo/config"
)

var (
	conf *config.Config
)

// ServeHTTP is used to serve the content of GIT API.
func ServeHTTP(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	conf = c

	switch r.Method {
	case "POST":
		return POST(w, r)
	default:
		return http.StatusMethodNotAllowed, errors.New("Invalid method.")
	}
}

package git

import (
	"net/http"

	"github.com/hacdias/caddy-hugo/config"
)

// ServeHTTP is used to serve the content of GIT API.
func ServeHTTP(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	if r.Method != http.MethodPost {
		return http.StatusNotImplemented, nil
	}

	return POST(w, r, c)
}

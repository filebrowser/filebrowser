package git

import (
	"net/http"

	"github.com/hacdias/caddy-hugo/config"
)

// POST handles the POST method on GIT page which is only an API.
func POST(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
	return http.StatusOK, nil
}

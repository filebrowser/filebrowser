package browse

import (
	"net/http"
	"os"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
)

// DELETE handles the delete requests on browse pages
func DELETE(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Remove both beginning and trailing slashes
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	path = c.Path + path

	// Check if the file or directory exists
	if stat, err := os.Stat(path); err == nil {
		var err error
		// If it's dir, remove all of the content inside
		if stat.IsDir() {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}

		// Check for errors
		if err != nil {
			w.Write([]byte(err.Error()))
			return 500, err
		}
	} else {
		return 404, nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
	return 200, nil
}

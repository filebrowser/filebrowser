package file

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/internal/config"
)

// NewDir makes a new directory
func NewDir(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	path := strings.Replace(r.URL.Path, c.BaseURL, c.PathScope, 1)
	path = filepath.Clean(path)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		switch {
		case os.IsPermission(err):
			return http.StatusForbidden, err
		case os.IsExist(err):
			return http.StatusConflict, err
		default:
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusCreated, nil
}

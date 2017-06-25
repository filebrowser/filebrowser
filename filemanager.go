// Package filemanager provides middleware for managing files in a directory
// when directory path is requested instead of a specific file. Based on browse
// middleware.
package filemanager

import (
	"net/http"

	"github.com/hacdias/filemanager"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// FileManager is an http.Handler that can show a file listing when
// directories in the given paths are specified.
type FileManager struct {
	Next    httpserver.Handler
	Configs []*filemanager.FileManager
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
func (f FileManager) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	for i := range f.Configs {
		// Checks if this Path should be handled by File Manager.
		if !httpserver.Path(r.URL.Path).Matches(f.Configs[i].BaseURL) {
			continue
		}

		return f.Configs[i].ServeHTTP(w, r)
	}

	return f.Next.ServeHTTP(w, r)
}

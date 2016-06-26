//go:generate go get github.com/jteeuwen/go-bindata
//go:generate go install github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -debug -pkg assets -prefix "assets" -o internal/assets/binary.go assets/...

// Package filemanager provides middleware for managing files in a directory
// when directory path is requested instead of a specific file. Based on browse
// middleware.
package filemanager

import (
	"net/http"
	"strings"

	a "github.com/hacdias/caddy-filemanager/internal/assets"
	"github.com/hacdias/caddy-filemanager/internal/config"
	"github.com/hacdias/caddy-filemanager/internal/file"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// FileManager is an http.Handler that can show a file listing when
// directories in the given paths are specified.
type FileManager struct {
	Next    httpserver.Handler
	Configs []config.Config
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
func (f FileManager) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	var (
		c      *config.Config
		fi     *file.Info
		code   int
		err    error
		assets bool
	)

	for i := range f.Configs {
		if httpserver.Path(r.URL.Path).Matches(f.Configs[i].BaseURL) {
			c = &f.Configs[i]
			assets = httpserver.Path(r.URL.Path).Matches(c.BaseURL + a.BaseURL)

			if r.Method != http.MethodPost && !assets {
				fi, code, err = file.GetInfo(r.URL, c)
				if err != nil {
					return code, err
				}

				if fi.IsDir && !strings.HasSuffix(r.URL.Path, "/") {
					http.Redirect(w, r, r.URL.Path+"/", http.StatusTemporaryRedirect)
					return 0, nil
				}
			}

			// Route the request depending on the HTTP Method
			switch r.Method {
			case http.MethodGet:
				// Read and show directory or file
				if assets {
					return a.ServeAssets(w, r, c)
				}

				if !fi.IsDir {
					query := r.URL.Query()
					if val, ok := query["raw"]; ok && val[0] == "true" {
						return fi.ServeRawFile(w, r, c)
					}

					if val, ok := query["download"]; ok && val[0] == "true" {
						w.Header().Set("Content-Disposition", "attachment; filename="+fi.Name)
						return fi.ServeRawFile(w, r, c)
					}
				}

				return fi.ServeAsHTML(w, r, c)
			case http.MethodPost:
				// Upload a new file
				if r.Header.Get("Upload") == "true" {
					return file.Upload(w, r, c)
				}
				// Search and git commands
				if r.Header.Get("Search") == "true" {
					// TODO: search and git commands
				}
				// Creates a new folder
				// TODO: not implemented on frontend
				return file.NewDir(w, r, c)
			case http.MethodDelete:
				// Delete a file or a directory
				return fi.Delete()
			case http.MethodPatch:
				// Rename a file or directory
				return fi.Rename(w, r)
			default:
				return http.StatusNotImplemented, nil
			}
		}
	}

	return f.Next.ServeHTTP(w, r)
}

//go:generate go get github.com/jteeuwen/go-bindata
//go:generate go install github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -debug -pkg assets -prefix "assets" -o internal/assets/binary.go assets/...

// Package filemanager provides middleware for managing files in a directory
// when directory path is requested instead of a specific file. Based on browse
// middleware.
package filemanager

import (
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
				fi, code, err = GetFileInfo(r.URL, c)
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
					return Upload(w, r, c)
				}
				return NewFolder(w, r, c)
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

// ErrorToHTTPCode gets the respective HTTP code for an error
func ErrorToHTTPCode(err error) int {
	switch {
	case os.IsPermission(err):
		return http.StatusForbidden
	case os.IsNotExist(err):
		return http.StatusNotFound
	case os.IsExist(err):
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}

// Upload is used to handle the upload requests to the server
func Upload(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Parse the multipart form in the request
	err := r.ParseMultipartForm(100000)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError, err
	}

	// For each file header in the multipart form
	for _, headers := range r.MultipartForm.File {
		// Handle each file
		for _, header := range headers {
			// Open the first file
			var src multipart.File
			if src, err = header.Open(); nil != err {
				return http.StatusInternalServerError, err
			}

			filename := strings.Replace(r.URL.Path, c.BaseURL, c.PathScope, 1)
			filename = filename + header.Filename
			filename = filepath.Clean(filename)

			// Create the file
			var dst *os.File
			if dst, err = os.Create(filename); nil != err {
				if os.IsExist(err) {
					return http.StatusConflict, err
				}
				return http.StatusInternalServerError, err
			}

			// Copy the file content
			if _, err = io.Copy(dst, src); nil != err {
				return http.StatusInternalServerError, err
			}

			defer dst.Close()
		}
	}

	return http.StatusOK, nil
}

// NewFolder makes a new directory
func NewFolder(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
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

//go:generate go get github.com/jteeuwen/go-bindata
//go:generate go install github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -debug -pkg filemanager -prefix "assets" -o binary.go assets/...
// TODO: remove debug from the comment

// Package filemanager provides middleware for managing files in a directory
// when directory path is requested instead of a specific file. Based on browse
// middleware.
package filemanager

import (
	"net/http"
	"os"
	"strings"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// FileManager is an http.Handler that can show a file listing when
// directories in the given paths are specified.
type FileManager struct {
	Next          httpserver.Handler
	Configs       []Config
	IgnoreIndexes bool
}

// Config is a configuration for browsing in a particular path.
type Config struct {
	PathScope  string
	Root       http.FileSystem
	BaseURL    string
	StyleSheet string
	Variables  interface{}
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
// If so, control is handed over to ServeListing.
func (f FileManager) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	var c *Config
	var file *InfoRequest

	// Check if there is a FileManager configuration to match the path
	for i := range f.Configs {
		if httpserver.Path(r.URL.Path).Matches(f.Configs[i].BaseURL) {
			c = &f.Configs[i]

			// Serve assets
			if httpserver.Path(r.URL.Path).Matches(c.BaseURL + "/_filemanagerinternal") {
				return ServeAssets(w, r, c)
			}

			// Gets the file path to be used within c.Root
			filepath := strings.Replace(r.URL.Path, c.BaseURL, "", 1)

			if r.Method != http.MethodPost {
				file = GetFileInfo(filepath, c)
				if file.Err != nil {
					defer file.File.Close()
					return file.Code, file.Err
				}

				if file.Info.IsDir() && !strings.HasSuffix(r.URL.Path, "/") {
					http.Redirect(w, r, r.URL.Path+"/", http.StatusTemporaryRedirect)
					return 0, nil
				}
			}

			// Route the request depending on the HTTP Method
			switch r.Method {
			case http.MethodGet:
				// Read and show directory or file
				if file.Info.IsDir() {
					return f.ServeListing(w, r, file.File, c)
				}
				return f.ServeSingleFile(w, r, file, c)
			case http.MethodPost:
				// Create new file or directory

				return http.StatusOK, nil
			case http.MethodDelete:
				// Delete a file or a directory
				return Delete(filepath, file.Info)
			case http.MethodPut:
				// Update/Modify a directory or file

				return http.StatusOK, nil
			case http.MethodPatch:
				// Rename a file or directory

				return http.StatusOK, nil
			default:
				return http.StatusNotImplemented, nil
			}
		}
	}
	return f.Next.ServeHTTP(w, r)
}

// InfoRequest is the information given by a GetFileInfo function
type InfoRequest struct {
	Info os.FileInfo
	File http.File
	Path string
	Code int
	Err  error
}

// GetFileInfo gets the file information and, in case of error, returns the
// respective HTTP error code
func GetFileInfo(path string, c *Config) *InfoRequest {
	request := &InfoRequest{Path: path}
	request.File, request.Err = c.Root.Open(path)
	if request.Err != nil {
		switch {
		case os.IsPermission(request.Err):
			request.Code = http.StatusForbidden
		case os.IsExist(request.Err):
			request.Code = http.StatusNotFound
		default:
			request.Code = http.StatusInternalServerError
		}

		return request
	}

	request.Info, request.Err = request.File.Stat()

	if request.Err != nil {
		switch {
		case os.IsPermission(request.Err):
			request.Code = http.StatusForbidden
		case os.IsExist(request.Err):
			request.Code = http.StatusGone
		default:
			request.Code = http.StatusInternalServerError
		}
	}

	return request
}

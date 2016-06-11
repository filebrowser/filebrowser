//go:generate go get github.com/jteeuwen/go-bindata
//go:generate go install github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -pkg filemanager -prefix "assets" -o binary.go assets/...

// Package filemanager provides middleware for managing files in a directory
// when directory path is requested instead of a specific file. Based on browse
// middleware.
package filemanager

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

const assetsURL = "/_filemanagerinternal"

// FileManager is an http.Handler that can show a file listing when
// directories in the given paths are specified.
type FileManager struct {
	Next    httpserver.Handler
	Configs []Config
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
func (f FileManager) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	var (
		c      *Config
		fi     *FileInfo
		code   int
		err    error
		assets bool
	)

	for i := range f.Configs {
		if httpserver.Path(r.URL.Path).Matches(f.Configs[i].BaseURL) {
			c = &f.Configs[i]
			assets = httpserver.Path(r.URL.Path).Matches(c.BaseURL + assetsURL)

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
					return ServeAssets(w, r, c)
				}

				if !fi.IsDir {
					query := r.URL.Query()
					if val, ok := query["raw"]; ok && val[0] == "true" {
						return f.Next.ServeHTTP(w, r)
					}

					if val, ok := query["download"]; ok && val[0] == "true" {
						w.Header().Set("Content-Disposition", "attachment; filename="+fi.Name)
						return f.Next.ServeHTTP(w, r)
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
	case os.IsExist(err):
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}

// ServeAssets provides the needed assets for the front-end
func ServeAssets(w http.ResponseWriter, r *http.Request, c *Config) (int, error) {
	// gets the filename to be used with Assets function
	filename := strings.Replace(r.URL.Path, c.BaseURL+assetsURL, "public", 1)
	file, err := Asset(filename)
	if err != nil {
		return http.StatusNotFound, nil
	}

	// Get the file extension and its mimetype
	extension := filepath.Ext(filename)
	mediatype := mime.TypeByExtension(extension)

	// Write the header with the Content-Type and write the file
	// content to the buffer
	w.Header().Set("Content-Type", mediatype)
	w.Write(file)
	return 200, nil
}

// Upload is used to handle the upload requests to the server
func Upload(w http.ResponseWriter, r *http.Request, c *Config) (int, error) {
	// Parse the multipart form in the request
	err := r.ParseMultipartForm(100000)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// For each file header in the multipart form
	for _, fheaders := range r.MultipartForm.File {
		// Handle each file
		for _, hdr := range fheaders {
			// Open the first file
			var infile multipart.File
			if infile, err = hdr.Open(); nil != err {
				return http.StatusInternalServerError, err
			}

			filename := strings.Replace(r.URL.Path, c.BaseURL, c.PathScope, 1)
			filename = filepath.Clean(filename)

			// Create the file
			var outfile *os.File
			if outfile, err = os.Create(filename); nil != err {
				if os.IsExist(err) {
					return http.StatusConflict, err
				}
				return http.StatusInternalServerError, err
			}

			// Copy the file content
			if _, err = io.Copy(outfile, infile); nil != err {
				return http.StatusInternalServerError, err
			}

			defer outfile.Close()
		}
	}

	return http.StatusOK, nil
}

// NewFolder makes a new directory
func NewFolder(w http.ResponseWriter, r *http.Request, c *Config) (int, error) {
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

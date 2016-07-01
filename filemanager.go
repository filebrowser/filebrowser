//go:generate go get github.com/jteeuwen/go-bindata
//go:generate go install github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -pkg assets -ignore .jsbeautifyrc -prefix "assets/embed" -o assets/binary.go assets/embed/...

// Package filemanager provides middleware for managing files in a directory
// when directory path is requested instead of a specific file. Based on browse
// middleware.
package filemanager

import (
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/assets"
	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/directory"
	"github.com/hacdias/caddy-filemanager/page"
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
		c           *config.Config
		fi          *directory.Info
		code        int
		err         error
		serveAssets bool
	)

	for i := range f.Configs {
		if httpserver.Path(r.URL.Path).Matches(f.Configs[i].BaseURL) {
			c = &f.Configs[i]
			serveAssets = httpserver.Path(r.URL.Path).Matches(c.BaseURL + assets.BaseURL)

			if r.Method != http.MethodPost && !serveAssets {
				fi, code, err = directory.GetInfo(r.URL, c)
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
				if serveAssets {
					return assets.Serve(w, r, c)
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
			case http.MethodPut:
				if fi.IsDir {
					return http.StatusNotAcceptable, nil
				}
				// Update a file
				return fi.Update(w, r, c)
			case http.MethodPost:
				// Upload a new file
				if r.Header.Get("Upload") == "true" {
					return upload(w, r, c)
				}
				// Search and git commands
				if r.Header.Get("Search") == "true" {
					// TODO: search commands
				}
				// VCS commands
				if r.Header.Get("Command") != "" {
					return vcsCommand(w, r, c)
				}
				// Creates a new folder
				return newDirectory(w, r, c)
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

// upload is used to handle the upload requests to the server
func upload(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
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

// newDirectory makes a new directory
func newDirectory(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	filename := r.Header.Get("Filename")

	if filename == "" {
		return http.StatusBadRequest, nil
	}

	path := strings.Replace(r.URL.Path, c.BaseURL, c.PathScope, 1) + filename
	path = filepath.Clean(path)
	extension := filepath.Ext(path)

	var err error

	if extension == "" {
		err = os.MkdirAll(path, 0755)
	} else {
		err = ioutil.WriteFile(path, []byte(""), 0755)
	}

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

// vcsCommand handles the requests for VCS related commands: git, svn and mercurial
func vcsCommand(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	command := strings.Split(r.Header.Get("command"), " ")

	// Check if the command is for git, mercurial or svn
	if command[0] != "git" && command[0] != "hg" && command[0] != "svn" {
		return http.StatusForbidden, nil
	}

	// Check if the program is talled is installed on the computer
	if _, err := exec.LookPath(command[0]); err != nil {
		return http.StatusNotImplemented, nil
	}

	path := strings.Replace(r.URL.Path, c.BaseURL, c.PathScope, 1)
	path = filepath.Clean(path)

	cmd := exec.Command(command[0], command[1:len(command)]...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()

	if err != nil {
		return http.StatusInternalServerError, err
	}

	page := &page.Page{Info: &page.Info{Data: string(output)}}
	return page.PrintAsJSON(w)
}

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
	"github.com/hacdias/caddy-filemanager/errors"
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
		user        *config.UserConfig
	)

	for i := range f.Configs {
		if httpserver.Path(r.URL.Path).Matches(f.Configs[i].BaseURL) {
			c = &f.Configs[i]
			serveAssets = httpserver.Path(r.URL.Path).Matches(c.BaseURL + assets.BaseURL)

			// Set the current User
			username, _, _ := r.BasicAuth()

			if _, ok := c.Users[username]; ok {
				user = c.Users[username]
			}

			if r.Method != http.MethodPost && !serveAssets {
				fi, code, err = directory.GetInfo(r.URL, c, user)
				if err != nil {
					if r.Method == http.MethodGet {
						return errors.PrintHTML(w, code, err)
					}
					return code, err
				}

				if fi.IsDir && !strings.HasSuffix(r.URL.Path, "/") {
					http.Redirect(w, r, c.AddrPath+r.URL.Path+"/", http.StatusTemporaryRedirect)
					return 0, nil
				}
			}

			// Secure agains CSRF attacks
			if r.Method != http.MethodGet {
				if !c.CheckToken(r) {
					return http.StatusForbidden, nil
				}
			}

			// Route the request depending on the HTTP Method
			switch r.Method {
			case http.MethodGet:
				// Read and show directory or file
				if serveAssets {
					return assets.Serve(w, r, c)
				}

				// Generate anti security token
				c.GenerateToken()

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

				code, err := fi.ServeAsHTML(w, r, c, user)
				if err != nil {
					return errors.PrintHTML(w, code, err)
				}
				return code, err
			case http.MethodPut:
				if fi.IsDir {
					return http.StatusNotAcceptable, nil
				}

				if !user.AllowEdit {
					return http.StatusUnauthorized, nil
				}

				// Update a file
				return fi.Update(w, r, c, user)
			case http.MethodPost:
				// Upload a new file
				if r.Header.Get("Upload") == "true" {
					if !user.AllowNew {
						return http.StatusUnauthorized, nil
					}

					return upload(w, r, c)
				}
				// Search and git commands
				if r.Header.Get("Search") == "true" {
					// TODO: search commands
				}
				// VCS commands
				if r.Header.Get("Command") != "" {
					if !user.AllowCommands {
						return http.StatusUnauthorized, nil
					}

					return command(w, r, c, user)
				}
				// Creates a new folder
				return newDirectory(w, r, c)
			case http.MethodDelete:
				if !user.AllowEdit {
					return http.StatusUnauthorized, nil
				}

				// Delete a file or a directory
				return fi.Delete()
			case http.MethodPatch:
				if !user.AllowEdit {
					return http.StatusUnauthorized, nil
				}

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

// command handles the requests for VCS related commands: git, svn and mercurial
func command(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.UserConfig) (int, error) {
	command := strings.Split(r.Header.Get("command"), " ")

	// Check if the command is allowed
	mayContinue := false

	for _, cmd := range u.Commands {
		if cmd == command[0] {
			mayContinue = true
		}
	}

	if !mayContinue {
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

//go:generate go get github.com/jteeuwen/go-bindata
//go:generate go install github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -pkg assets -ignore .jsbeautifyrc -prefix "assets/embed" -o assets/binary.go assets/embed/...

// Package filemanager provides middleware for managing files in a directory
// when directory path is requested instead of a specific file. Based on browse
// middleware.
package filemanager

import (
	e "errors"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/assets"
	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/errors"
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
		c    *config.Config
		fi   *FileInfo
		code int
		err  error
		user *config.User
	)

	for i := range f.Configs {
		if httpserver.Path(r.URL.Path).Matches(f.Configs[i].BaseURL) {
			c = &f.Configs[i]

			if r.Method == http.MethodGet && httpserver.Path(r.URL.Path).Matches(c.BaseURL+assets.BaseURL) {
				return assets.Serve(w, r, c)
			}

			username, _, _ := r.BasicAuth()

			if _, ok := c.Users[username]; ok {
				user = c.Users[username]
			} else {
				user = c.User
			}

			if strings.HasPrefix(r.URL.Path, c.WebDavURL) {
				if !user.Allowed(strings.TrimPrefix(r.URL.Path, c.WebDavURL)) {
					return http.StatusForbidden, nil
				}

				switch r.Method {
				case "PROPPATCH", "MOVE", "PATCH", "PUT", "DELETE":
					if !user.AllowEdit {
						return http.StatusForbidden, nil
					}
				case "MKCOL", "COPY":
					if !user.AllowNew {
						return http.StatusForbidden, nil
					}
				}

				if r.Method == http.MethodPut {
					_, err = fi.Update(w, r, c, user)
					if err != nil {
						return http.StatusInternalServerError, err
					}
				}

				c.Handler.ServeHTTP(w, r)
				return 0, nil
			}

			if !user.Allowed(strings.TrimPrefix(r.URL.Path, c.BaseURL)) {
				if r.Method == http.MethodGet {
					return errors.PrintHTML(
						w,
						http.StatusForbidden,
						e.New("You don't have permission to access this page."),
					)
				}

				return http.StatusForbidden, nil
			}

			if r.Method == http.MethodGet {
				// Gets the information of the directory/file
				fi, code, err = GetInfo(r.URL, c, user)
				if err != nil {
					if r.Method == http.MethodGet {
						return errors.PrintHTML(w, code, err)
					}
					return code, err
				}

				// If it's a dir and the path doesn't end with a trailing slash,
				// redirect the user.
				if fi.IsDir() && !strings.HasSuffix(r.URL.Path, "/") {
					http.Redirect(w, r, c.AddrPath+r.URL.Path+"/", http.StatusTemporaryRedirect)
					return 0, nil
				}

				// Generate anti security token.
				c.GenerateToken()

				if !fi.IsDir() {
					query := r.URL.Query()
					if val, ok := query["raw"]; ok && val[0] == "true" {
						r.URL.Path = strings.Replace(r.URL.Path, c.BaseURL, c.WebDavURL, 1)
						c.Handler.ServeHTTP(w, r)
						return 0, nil
					}

					if val, ok := query["download"]; ok && val[0] == "true" {
						w.Header().Set("Content-Disposition", "attachment; filename="+fi.Name())
						r.URL.Path = strings.Replace(r.URL.Path, c.BaseURL, c.WebDavURL, 1)
						c.Handler.ServeHTTP(w, r)
						return 0, nil
					}
				}

				code, err := fi.ServeHTTP(w, r, c, user)
				if err != nil {
					return errors.PrintHTML(w, code, err)
				}
				return code, err
			}

			if r.Method == http.MethodPost {
				// TODO: How to exclude web dav clients? :/
				// Security measures against CSRF attacks.
				if !c.CheckToken(r) {
					return http.StatusForbidden, nil
				}

				/* TODO: search commands. USE PROPFIND?
				// Search and git commands.
				if r.Header.Get("Search") == "true" {

				} */

				// VCS commands.
				if r.Header.Get("Command") != "" {
					if !user.AllowCommands {
						return http.StatusUnauthorized, nil
					}

					return command(w, r, c, user)
				}
			}

			return http.StatusNotImplemented, nil
		}
	}

	return f.Next.ServeHTTP(w, r)
}

// command handles the requests for VCS related commands: git, svn and mercurial
func command(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.User) (int, error) {
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

	path := strings.Replace(r.URL.Path, c.BaseURL, c.Scope, 1)
	path = filepath.Clean(path)

	cmd := exec.Command(command[0], command[1:len(command)]...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()

	if err != nil {
		return http.StatusInternalServerError, err
	}

	p := &page{pageInfo: &pageInfo{Data: string(output)}}
	return p.PrintAsJSON(w)
}

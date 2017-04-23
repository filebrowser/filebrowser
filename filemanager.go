// Package filemanager provides middleware for managing files in a directory
// when directory path is requested instead of a specific file. Based on browse
// middleware.
package filemanager

import (
	e "errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/assets"
	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/file"
	"github.com/hacdias/caddy-filemanager/handlers"
	"github.com/hacdias/caddy-filemanager/page"
	"github.com/hacdias/caddy-filemanager/wrapper"
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
		fi   *file.Info
		code int
		err  error
		user *config.User
	)

	for i := range f.Configs {
		// Checks if this Path should be handled by File Manager.
		if !httpserver.Path(r.URL.Path).Matches(f.Configs[i].BaseURL) {
			continue
		}

		c = &f.Configs[i]

		// Checks if the URL matches the Assets URL. Returns the asset if the
		// method is GET and Status Forbidden otherwise.
		if httpserver.Path(r.URL.Path).Matches(c.BaseURL + assets.BaseURL) {
			if r.Method == http.MethodGet {
				return assets.Serve(w, r, c)
			}

			return http.StatusForbidden, nil
		}

		// Obtains the user. See https://github.com/mholt/caddy/blob/master/caddyhttp/basicauth/basicauth.go#L66
		username, _ := r.Context().Value(httpserver.RemoteUserCtxKey).(string)
		if _, ok := c.Users[username]; ok {
			user = c.Users[username]
		} else {
			user = c.User
		}

		// Checks if the request URL is for the WebDav server
		if httpserver.Path(r.URL.Path).Matches(c.WebDavURL) {
			// Checks for user permissions relatively to this PATH
			if !user.Allowed(strings.TrimPrefix(r.URL.Path, c.WebDavURL)) {
				return http.StatusForbidden, nil
			}

			switch r.Method {
			case "GET", "HEAD":
				// Excerpt from RFC4918, section 9.4:
				//
				// 		GET, when applied to a collection, may return the contents of an
				//		"index.html" resource, a human-readable view of the contents of
				//		the collection, or something else altogether.
				//
				// It was decided on https://github.com/hacdias/caddy-filemanager/issues/85
				// that GET, for collections, will return the same as PROPFIND method.
				path := strings.Replace(r.URL.Path, c.WebDavURL, "", 1)
				path = user.Scope + "/" + path
				path = filepath.Clean(path)

				var i os.FileInfo
				i, err = os.Stat(path)
				if err != nil {
					// Is there any error? WebDav will handle it... no worries.
					break
				}

				if i.IsDir() {
					r.Method = "PROPFIND"

					if r.Method == "HEAD" {
						w = wrapper.NewResponseWriterNoBody(w)
					}
				}
			case "PROPPATCH", "MOVE", "PATCH", "PUT", "DELETE":
				if !user.AllowEdit {
					return http.StatusForbidden, nil
				}
			case "MKCOL", "COPY":
				if !user.AllowNew {
					return http.StatusForbidden, nil
				}
			}

			// Preprocess the PUT request if it's the case
			if r.Method == http.MethodPut {
				if err = c.BeforeSave(r, c, user); err != nil {
					return http.StatusInternalServerError, err
				}

				if handlers.PreProccessPUT(w, r, c, user) != nil {
					return http.StatusInternalServerError, err
				}
			}

			c.Handler.ServeHTTP(w, r)
			if err = c.AfterSave(r, c, user); err != nil {
				return http.StatusInternalServerError, err
			}

			return 0, nil
		}

		w.Header().Set("x-frame-options", "SAMEORIGIN")
		w.Header().Set("x-content-type", "nosniff")
		w.Header().Set("x-xss-protection", "1; mode=block")

		// Checks if the User is allowed to access this file
		if !user.Allowed(strings.TrimPrefix(r.URL.Path, c.BaseURL)) {
			if r.Method == http.MethodGet {
				return page.PrintErrorHTML(
					w, http.StatusForbidden,
					e.New("You don't have permission to access this page."),
				)
			}

			return http.StatusForbidden, nil
		}

		if r.URL.Query().Get("search") != "" {
			return handlers.Search(w, r, c, user)
		}

		if r.URL.Query().Get("command") != "" {
			return handlers.Command(w, r, c, user)
		}

		if r.Method == http.MethodGet {
			// Gets the information of the directory/file
			fi, code, err = file.GetInfo(r.URL, c, user)
			if err != nil {
				if r.Method == http.MethodGet {
					return page.PrintErrorHTML(w, code, err)
				}
				return code, err
			}

			// If it's a dir and the path doesn't end with a trailing slash,
			// redirect the user.
			if fi.IsDir && !strings.HasSuffix(r.URL.Path, "/") {
				http.Redirect(w, r, c.PrefixURL+r.URL.Path+"/", http.StatusTemporaryRedirect)
				return 0, nil
			}

			switch {
			case r.URL.Query().Get("download") != "":
				code, err = handlers.Download(w, r, c, fi)
			case r.URL.Query().Get("raw") == "true" && !fi.IsDir:
				http.ServeFile(w, r, fi.Path)
				code, err = 0, nil
			case !fi.IsDir && r.URL.Query().Get("checksum") != "":
				code, err = handlers.Checksum(w, r, c, fi)
			case fi.IsDir:
				code, err = handlers.ServeListing(w, r, c, user, fi)
			default:
				code, err = handlers.ServeSingle(w, r, c, user, fi)
			}

			if err != nil {
				code, err = page.PrintErrorHTML(w, code, err)
			}

			return code, err
		}

		return http.StatusNotImplemented, nil

	}

	return f.Next.ServeHTTP(w, r)
}

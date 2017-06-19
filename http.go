package filemanager

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// ServeHTTP starts FileManager.
func (c *FileManager) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	var (
		fi   *fileInfo
		user *User
		code int
		err  error
	)
	/* TODO: readd this
	// Checks if the URL matches the Assets URL. Returns the asset if the
	// method is GET and Status Forbidden otherwise.
	if strings.HasPrefix(r.URL.Path, c.BaseURL+assets.BaseURL) {
		if r.Method == http.MethodGet {
			return assets.Serve(w, r, c)
		}

		return http.StatusForbidden, nil
	} */

	// Obtains the user.
	username, _, _ := r.BasicAuth()
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
					w = NewResponseWriterNoBody(w)
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

			if c.preProccessPUT(w, r, user) != nil {
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
			return printError(
				w, http.StatusForbidden,
				errors.New("You don't have permission to access this page"),
			)
		}

		return http.StatusForbidden, nil
	}

	if r.URL.Query().Get("search") != "" {
		return c.search(w, r, user)
	}

	if r.URL.Query().Get("command") != "" {
		return c.command(w, r, user)
	}

	if r.Method == http.MethodGet {
		// Gets the information of the directory/file
		fi, err = getFileInfo(r.URL, c, user)
		code = errorToHTTPCode(err, false)
		if err != nil {
			if r.Method == http.MethodGet {
				return printError(w, code, err)
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
			code, err = c.download(w, r, fi)
		case r.URL.Query().Get("raw") == "true" && !fi.IsDir:
			http.ServeFile(w, r, fi.Path)
			code, err = 0, nil
		case !fi.IsDir && r.URL.Query().Get("checksum") != "":
			code, err = c.checksum(w, r, fi)
		case fi.IsDir:
			code, err = c.serveListing(w, r, user, fi)
		default:
			code, err = c.serveSingle(w, r, user, fi)
		}

		if err != nil {
			code, err = printError(w, code, err)
		}

		return code, err
	}

	return http.StatusNotImplemented, nil
}

package filemanager

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func matchURL(first, second string) bool {
	first = strings.ToLower(first)
	second = strings.ToLower(second)

	return strings.HasPrefix(first, second)
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
func (c *FileManager) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	var (
		fi   *FileInfo
		code int
		err  error
		user *user
	)

	// Checks if the URL matches the Assets URL. Returns the asset if the
	// method is GET and Status Forbidden otherwise.
	if matchURL(r.URL.Path, c.BaseURL+AssetsURL) {
		if r.Method == http.MethodGet {
			return serveAssets(w, r, c)
		}

		return http.StatusForbidden, nil
	}

	username, _, _ := r.BasicAuth()
	if _, ok := c.Users[username]; ok {
		user = c.Users[username]
	} else {
		user = c.user
	}

	// Checks if the request URL is for the WebDav server
	if matchURL(r.URL.Path, c.WebDavURL) {
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
			path = user.scope + "/" + path
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
					w = newResponseWriterNoBody(w)
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

			if put(w, r, c, user) != nil {
				return http.StatusInternalServerError, err
			}
		}

		c.handler.ServeHTTP(w, r)
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
			return htmlError(
				w, http.StatusForbidden,
				errors.New("You don't have permission to access this page"),
			)
		}

		return http.StatusForbidden, nil
	}

	if r.URL.Query().Get("search") != "" {
		return search(w, r, c, user)
	}

	if r.URL.Query().Get("command") != "" {
		return command(w, r, c, user)
	}

	if r.Method == http.MethodGet {
		// Gets the information of the directory/file
		fi, err = GetInfo(r.URL, c, user)
		if err != nil {
			if r.Method == http.MethodGet {
				return htmlError(w, code, err)
			}
			code = errorToHTTP(err, false)
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
			code, err = download(w, r, fi)
		case r.URL.Query().Get("raw") == "true" && !fi.IsDir:
			http.ServeFile(w, r, fi.Path)
			code, err = 0, nil
		case !fi.IsDir && r.URL.Query().Get("checksum") != "":
			code, err = checksum(w, r, fi)
		case fi.IsDir:
			code, err = serveListing(w, r, c, user, fi)
		default:
			code, err = serveSingle(w, r, c, user, fi)
		}

		if err != nil {
			code, err = htmlError(w, code, err)
		}

		return code, err
	}

	return http.StatusNotImplemented, nil
}

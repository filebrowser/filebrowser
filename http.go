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
func (m *FileManager) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	var (
		u    *user
		code int
		err  error
	)

	// Checks if the URL matches the Assets URL. Returns the asset if the
	// method is GET and Status Forbidden otherwise.
	if matchURL(r.URL.Path, m.baseURL+assetsURL) {
		if r.Method == http.MethodGet {
			return serveAssets(w, r, m)
		}

		return http.StatusForbidden, nil
	}

	username, _, _ := r.BasicAuth()
	if _, ok := m.Users[username]; ok {
		u = m.Users[username]
	} else {
		u = m.user
	}

	// Checks if the request URL is for the WebDav server
	if matchURL(r.URL.Path, m.webDavURL) {
		return serveWebDAV(w, r, m, u)
	}

	w.Header().Set("x-frame-options", "SAMEORIGIN")
	w.Header().Set("x-content-type", "nosniff")
	w.Header().Set("x-xss-protection", "1; mode=block")

	// Checks if the User is allowed to access this file
	if !u.Allowed(strings.TrimPrefix(r.URL.Path, m.baseURL)) {
		if r.Method == http.MethodGet {
			return htmlError(
				w, http.StatusForbidden,
				errors.New("You don't have permission to access this page"),
			)
		}

		return http.StatusForbidden, nil
	}

	if r.URL.Query().Get("search") != "" {
		return search(w, r, m, u)
	}

	if r.URL.Query().Get("command") != "" {
		return command(w, r, m, u)
	}

	if r.Method == http.MethodGet {
		var f *fileInfo

		// Obtains the information of the directory/file.
		f, err = getInfo(r.URL, m, u)
		if err != nil {
			if r.Method == http.MethodGet {
				return htmlError(w, code, err)
			}

			code = errorToHTTP(err, false)
			return code, err
		}

		// If it's a dir and the path doesn't end with a trailing slash,
		// redirect the user.
		if f.IsDir && !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, m.PrefixURL+r.URL.Path+"/", http.StatusTemporaryRedirect)
			return 0, nil
		}

		switch {
		case r.URL.Query().Get("download") != "":
			code, err = download(w, r, f)
		case r.URL.Query().Get("raw") == "true" && !f.IsDir:
			http.ServeFile(w, r, f.Path)
			code, err = 0, nil
		case !f.IsDir && r.URL.Query().Get("checksum") != "":
			code, err = checksum(w, r, f)
		case f.IsDir:
			code, err = serveListing(w, r, m, u, f)
		default:
			code, err = serveSingle(w, r, m, u, f)
		}

		if err != nil {
			code, err = htmlError(w, code, err)
		}

		return code, err
	}

	return http.StatusNotImplemented, nil
}

// serveWebDAV handles the webDAV route of the File Manager.
func serveWebDAV(w http.ResponseWriter, r *http.Request, m *FileManager, u *user) (int, error) {
	var err error

	// Checks for user permissions relatively to this path.
	if !u.Allowed(strings.TrimPrefix(r.URL.Path, m.webDavURL)) {
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
		path := strings.Replace(r.URL.Path, m.webDavURL, "", 1)
		path = u.scope + "/" + path
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
		if !u.AllowEdit {
			return http.StatusForbidden, nil
		}
	case "MKCOL", "COPY":
		if !u.AllowNew {
			return http.StatusForbidden, nil
		}
	}

	// Preprocess the PUT request if it's the case
	if r.Method == http.MethodPut {
		if err = m.BeforeSave(r, m, u); err != nil {
			return http.StatusInternalServerError, err
		}

		if put(w, r, m, u) != nil {
			return http.StatusInternalServerError, err
		}
	}

	m.handler.ServeHTTP(w, r)
	if err = m.AfterSave(r, m, u); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

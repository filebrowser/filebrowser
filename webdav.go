package filemanager

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// serveWebDAV handles the webDAV route of the File Manager.
func serveWebDAV(w http.ResponseWriter, r *http.Request, m *FileManager, u *User) (int, error) {
	var err error

	// Checks for user permissions relatively to this path.
	if !u.Allowed(strings.TrimPrefix(r.URL.Path, m.WebDavURL)) {
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
		path := strings.Replace(r.URL.Path, m.WebDavURL, "", 1)
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

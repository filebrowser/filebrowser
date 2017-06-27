package filemanager

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func serveHTTP(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var (
		code int
		err  error
	)

	// Checks if the URL contains the baseURL. If so, it strips it. Otherwise,
	// it throws an error.
	if p := strings.TrimPrefix(r.URL.Path, c.fm.baseURL); len(p) < len(r.URL.Path) {
		r.URL.Path = p
	} else {
		return http.StatusNotFound, nil
	}

	// Checks if the URL matches the Assets URL. Returns the asset if the
	// method is GET and Status Forbidden otherwise.
	if matchURL(r.URL.Path, assetsURL) {
		if r.Method == http.MethodGet {
			return serveAssets(c, w, r)
		}

		return http.StatusForbidden, nil
	}

	username, _, _ := r.BasicAuth()
	if _, ok := c.fm.Users[username]; ok {
		c.us = c.fm.Users[username]
	} else {
		c.us = c.fm.User
	}

	// Checks if the request URL is for the WebDav server.
	if matchURL(r.URL.Path, c.fm.webDavURL) {
		return serveWebDAV(c, w, r)
	}

	w.Header().Set("x-frame-options", "SAMEORIGIN")
	w.Header().Set("x-content-type", "nosniff")
	w.Header().Set("x-xss-protection", "1; mode=block")

	// Checks if the User is allowed to access this file
	if !c.us.Allowed(r.URL.Path) {
		if r.Method == http.MethodGet {
			return htmlError(
				w, http.StatusForbidden,
				errors.New("You don't have permission to access this page"),
			)
		}

		return http.StatusForbidden, nil
	}

	if r.URL.Query().Get("search") != "" {
		return search(c, w, r)
	}

	if r.URL.Query().Get("command") != "" {
		return command(c, w, r)
	}

	if r.Method == http.MethodGet {
		var f *fileInfo

		// Obtains the information of the directory/file.
		f, err = getInfo(r.URL, c.fm, c.us)
		if err != nil {
			if r.Method == http.MethodGet {
				return htmlError(w, code, err)
			}

			code = errorToHTTP(err, false)
			return code, err
		}

		c.fi = f

		// If it's a dir and the path doesn't end with a trailing slash,
		// redirect the user.
		if f.IsDir && !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, c.fm.RootURL()+r.URL.Path+"/", http.StatusTemporaryRedirect)
			return 0, nil
		}

		switch {
		case r.URL.Query().Get("download") != "":
			code, err = download(c, w, r)
		case !f.IsDir && r.URL.Query().Get("checksum") != "":
			code, err = serveChecksum(c, w, r)
		case r.URL.Query().Get("raw") == "true" && !f.IsDir:
			http.ServeFile(w, r, f.Path)
			code, err = 0, nil
		case f.IsDir:
			code, err = serveListing(c, w, r)
		default:
			code, err = serveSingle(c, w, r)
		}

		if err != nil {
			code, err = htmlError(w, code, err)
		}

		return code, err
	}

	return http.StatusNotImplemented, nil
}

// serveWebDAV handles the webDAV route of the File Manager.
func serveWebDAV(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var err error

	// Checks for user permissions relatively to this path.
	if !c.us.Allowed(strings.TrimPrefix(r.URL.Path, c.fm.webDavURL)) {
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
		path := strings.Replace(r.URL.Path, c.fm.webDavURL, "", 1)
		path = c.us.scope + "/" + path
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
		if !c.us.AllowEdit {
			return http.StatusForbidden, nil
		}
	case "MKCOL", "COPY":
		if !c.us.AllowNew {
			return http.StatusForbidden, nil
		}
	}

	// Preprocess the PUT request if it's the case
	if r.Method == http.MethodPut {
		if err = c.fm.BeforeSave(r, c.fm, c.us); err != nil {
			return http.StatusInternalServerError, err
		}

		if put(c, w, r) != nil {
			return http.StatusInternalServerError, err
		}
	}

	c.fm.handler.ServeHTTP(w, r)
	if err = c.fm.AfterSave(r, c.fm, c.us); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

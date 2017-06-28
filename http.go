package filemanager

import (
	"errors"
	"net/http"
	"os"
	"strings"
)

// assetsURL is the url where static assets are served.
const assetsURL = "/_"

// requestContext contains the needed information to make handlers work.
type requestContext struct {
	us *User
	fm *FileManager
	fi *fileInfo
	pg *page
}

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
	if matchURL(r.URL.Path, assetsURL+"/") {
		if r.Method == http.MethodGet {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, assetsURL)
			c.fm.static.ServeHTTP(w, r)
			return 0, nil
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
			code, err = serveDownload(c, w, r)
		case !f.IsDir && r.URL.Query().Get("checksum") != "":
			code, err = serveChecksum(c, w, r)
		case r.URL.Query().Get("raw") == "true" && !f.IsDir:
			http.ServeFile(w, r, f.Path)
			code, err = 0, nil
		default:
			code, err = serveDefault(c, w, r)
		}

		if err != nil {
			code, err = htmlError(w, code, err)
		}

		return code, err
	}

	return http.StatusNotImplemented, nil
}

// responseWriterNoBody is a wrapper used to suprress the body of the response
// to a request. Mainly used for HEAD requests.
type responseWriterNoBody struct {
	http.ResponseWriter
}

// newResponseWriterNoBody creates a new responseWriterNoBody.
func newResponseWriterNoBody(w http.ResponseWriter) *responseWriterNoBody {
	return &responseWriterNoBody{w}
}

// Header executes the Header method from the http.ResponseWriter.
func (w responseWriterNoBody) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Write suprresses the body.
func (w responseWriterNoBody) Write(data []byte) (int, error) {
	return 0, nil
}

// WriteHeader writes the header to the http.ResponseWriter.
func (w responseWriterNoBody) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

// matchURL checks if the first URL matches the second.
func matchURL(first, second string) bool {
	first = strings.ToLower(first)
	second = strings.ToLower(second)

	return strings.HasPrefix(first, second)
}

// errorToHTTP converts errors to HTTP Status Code.
func errorToHTTP(err error, gone bool) int {
	switch {
	case os.IsPermission(err):
		return http.StatusForbidden
	case os.IsNotExist(err):
		if !gone {
			return http.StatusNotFound
		}

		return http.StatusGone
	case os.IsExist(err):
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}

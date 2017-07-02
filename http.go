package filemanager

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"strings"
)

// requestContext contains the needed information to make handlers work.
type requestContext struct {
	us *User
	fm *FileManager
	fi *file
}

// serveHTTP is the main entry point of this HTML application.
func serveHTTP(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// Checks if the URL contains the baseURL and strips it. Otherwise, it just
	// returns a 404 error because we're not supposed to be here!
	p := strings.TrimPrefix(r.URL.Path, c.fm.baseURL)

	if len(p) >= len(r.URL.Path) && c.fm.baseURL != "" {
		return http.StatusNotFound, nil
	}

	r.URL.Path = p

	// Check if this request is made to the service worker. If so,
	// pass it through a template to add the needed variables.
	if r.URL.Path == "/sw.js" {
		return renderFile(
			w,
			c.fm.assets.MustString(r.URL.Path),
			"application/javascript",
			c.fm.RootURL(),
		)
	}

	// Checks if this request is made to the static assets folder. If so, and
	// if it is a GET request, returns with the asset. Otherwise, returns
	// a status not implemented.
	if matchURL(r.URL.Path, "/static") {
		if r.Method != http.MethodGet {
			return http.StatusNotImplemented, nil
		}

		return staticHandler(c, w, r)
	}

	// Checks if this request is made to the API and directs to the
	// API handler if so.
	if matchURL(r.URL.Path, "/api") {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
		return serveAPI(c, w, r)
	}

	// Checks if this request is made to the base path /files. If so,
	// shows the index.html page.
	if matchURL(r.URL.Path, "/files") {
		w.Header().Set("x-frame-options", "SAMEORIGIN")
		w.Header().Set("x-content-type", "nosniff")
		w.Header().Set("x-xss-protection", "1; mode=block")

		return renderFile(
			w,
			c.fm.assets.MustString("index.html"),
			"text/html",
			c.fm.RootURL(),
		)
	}

	http.Redirect(w, r, c.fm.RootURL()+"/files"+r.URL.Path, http.StatusTemporaryRedirect)
	return 0, nil
}

// staticHandler handles the static assets path.
func staticHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path != "/static/manifest.json" {
		http.FileServer(c.fm.assets.HTTPBox()).ServeHTTP(w, r)
		return 0, nil
	}

	return renderFile(
		w,
		c.fm.assets.MustString(r.URL.Path),
		"application/json",
		c.fm.RootURL(),
	)
}

// serveChecksum calculates the hash of a file. Supports MD5, SHA1, SHA256 and SHA512.
func checksumHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	query := r.URL.Query().Get("algo")

	val, err := c.fi.Checksum(query)
	if err == errInvalidOption {
		return http.StatusBadRequest, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Write([]byte(val))
	return 0, nil
}

// renderFile renders a file using a template with some needed variables.
func renderFile(w http.ResponseWriter, file string, contentType string, baseURL string) (int, error) {
	tpl := template.Must(template.New("file").Parse(file))
	w.Header().Set("Content-Type", contentType+"; charset=utf-8")

	err := tpl.Execute(w, map[string]string{"BaseURL": baseURL})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

// renderJSON prints the JSON version of data to the browser.
func renderJSON(w http.ResponseWriter, data interface{}) (int, error) {
	marsh, err := json.Marshal(data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(marsh); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
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
	case err == nil:
		return http.StatusOK
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

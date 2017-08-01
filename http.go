package filemanager

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"strings"
)

// RequestContext contains the needed information to make handlers work.
type RequestContext struct {
	User *User
	FM   *FileManager
	FI   *file
	// On API handlers, Router is the APi handler we want.
	Router string
}

// serveHTTP is the main entry point of this HTML application.
func serveHTTP(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// Checks if the URL contains the baseURL and strips it. Otherwise, it just
	// returns a 404 error because we're not supposed to be here!
	p := strings.TrimPrefix(r.URL.Path, c.FM.BaseURL)

	if len(p) >= len(r.URL.Path) && c.FM.BaseURL != "" {
		return http.StatusNotFound, nil
	}

	r.URL.Path = p

	// Check if this request is made to the service worker. If so,
	// pass it through a template to add the needed variables.
	if r.URL.Path == "/sw.js" {
		return renderFile(
			w,
			c.FM.assets.MustString("sw.js"),
			"application/javascript",
			c,
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
		return apiHandler(c, w, r)
	}

	// Any other request should show the index.html file.
	w.Header().Set("x-frame-options", "SAMEORIGIN")
	w.Header().Set("x-content-type", "nosniff")
	w.Header().Set("x-xss-protection", "1; mode=block")

	return renderFile(
		w,
		c.FM.assets.MustString("index.html"),
		"text/html",
		c,
	)
}

// staticHandler handles the static assets path.
func staticHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path != "/static/manifest.json" {
		http.FileServer(c.FM.assets.HTTPBox()).ServeHTTP(w, r)
		return 0, nil
	}

	return renderFile(
		w,
		c.FM.assets.MustString("static/manifest.json"),
		"application/json",
		c,
	)
}

// apiHandler is the main entry point for the /api endpoint.
func apiHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path == "/auth/get" {
		return authHandler(c, w, r)
	}

	if r.URL.Path == "/auth/renew" {
		return renewAuthHandler(c, w, r)
	}

	valid, _ := validateAuth(c, r)
	if !valid {
		return http.StatusForbidden, nil
	}

	c.Router, r.URL.Path = splitURL(r.URL.Path)

	if !c.User.Allowed(r.URL.Path) {
		return http.StatusForbidden, nil
	}

	for p := range c.FM.Plugins {
		code, err := plugins[p].Handler.Before(c, w, r)
		if code != 0 || err != nil {
			return code, err
		}
	}

	if c.Router == "checksum" || c.Router == "download" {
		var err error
		c.FI, err = getInfo(r.URL, c.FM, c.User)
		if err != nil {
			return errorToHTTP(err, false), err
		}
	}

	var code int
	var err error

	switch c.Router {
	case "download":
		code, err = downloadHandler(c, w, r)
	case "checksum":
		code, err = checksumHandler(c, w, r)
	case "command":
		code, err = command(c, w, r)
	case "search":
		code, err = search(c, w, r)
	case "resource":
		code, err = resourceHandler(c, w, r)
	case "users":
		code, err = usersHandler(c, w, r)
	case "settings":
		code, err = settingsHandler(c, w, r)
	default:
		code = http.StatusNotFound
	}

	if code >= 300 || err != nil {
		return code, err
	}

	for p := range c.FM.Plugins {
		code, err := plugins[p].Handler.After(c, w, r)
		if code != 0 || err != nil {
			return code, err
		}
	}

	return code, err
}

// serveChecksum calculates the hash of a file. Supports MD5, SHA1, SHA256 and SHA512.
func checksumHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	query := r.URL.Query().Get("algo")

	val, err := c.FI.Checksum(query)
	if err == errInvalidOption {
		return http.StatusBadRequest, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Write([]byte(val))
	return 0, nil
}

// splitURL splits the path and returns everything that stands
// before the first slash and everything that goes after.
func splitURL(path string) (string, string) {
	if path == "" {
		return "", ""
	}

	path = strings.TrimPrefix(path, "/")

	i := strings.Index(path, "/")
	if i == -1 {
		return "", path
	}

	return path[0:i], path[i:]
}

// renderFile renders a file using a template with some needed variables.
func renderFile(w http.ResponseWriter, file string, contentType string, c *RequestContext) (int, error) {
	tpl := template.Must(template.New("file").Parse(file))
	w.Header().Set("Content-Type", contentType+"; charset=utf-8")

	var javascript = ""
	for name := range c.FM.Plugins {
		javascript += plugins[name].JavaScript + "\n"
	}

	err := tpl.Execute(w, map[string]interface{}{
		"BaseURL":    c.FM.RootURL(),
		"JavaScript": template.JS(javascript),
	})
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
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

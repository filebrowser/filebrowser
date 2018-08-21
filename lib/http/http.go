package http

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	fb "github.com/filebrowser/filebrowser/lib"
)

// Handler returns a function compatible with http.HandleFunc.
func Handler(m *fb.FileBrowser) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code, err := serve(&fb.Context{
			FileBrowser: m,
			User:        nil,
			File:        nil,
		}, w, r)

		if code >= 400 {
			w.WriteHeader(code)

			txt := http.StatusText(code)
			log.Printf("%v: %v %v\n", r.URL.Path, code, txt)
			w.Write([]byte(txt + "\n"))
		}

		if err != nil {
			log.Print(err)
		}
	})
}

// serve is the main entry point of this HTML application.
func serve(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	// Checks if the URL contains the baseURL and strips it. Otherwise, it just
	// returns a 404 fb.Error because we're not supposed to be here!
	p := strings.TrimPrefix(r.URL.Path, c.BaseURL)

	if len(p) >= len(r.URL.Path) && c.BaseURL != "" {
		return http.StatusNotFound, nil
	}

	r.URL.Path = p

	// Check if this request is made to the service worker. If so,
	// pass it through a template to add the needed variables.
	if r.URL.Path == "/sw.js" {
		return renderFile(c, w, "sw.js")
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

	// If it is a request to the preview and a static website generator is
	// active, build the preview.
	if strings.HasPrefix(r.URL.Path, "/preview") && c.StaticGen != nil {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/preview")
		return c.StaticGen.Preview(c, w, r)
	}

	if strings.HasPrefix(r.URL.Path, "/share/") {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/share/")
		return sharePage(c, w, r)
	}

	// Any other request should show the index.html file.
	w.Header().Set("x-frame-options", "SAMEORIGIN")
	w.Header().Set("x-content-type-options", "nosniff")
	w.Header().Set("x-xss-protection", "1; mode=block")

	return renderFile(c, w, "index.html")
}

// staticHandler handles the static assets path.
func staticHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path != "/static/manifest.json" {
		http.FileServer(c.Assets.HTTPBox()).ServeHTTP(w, r)
		return 0, nil
	}

	return renderFile(c, w, "static/manifest.json")
}

// apiHandler is the main entry point for the /api endpoint.
func apiHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
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

	if c.StaticGen != nil {
		// If we are using the 'magic url' for the settings,
		// we should redirect the request for the acutual path.
		if r.URL.Path == "/settings" {
			r.URL.Path = c.StaticGen.SettingsPath()
		}

		// Executes the Static website generator hook.
		code, err := c.StaticGen.Hook(c, w, r)
		if code != 0 || err != nil {
			return code, err
		}
	}

	if c.Router == "checksum" || c.Router == "download" || c.Router == "subtitle" || c.Router == "subtitles" {
		var err error
		c.File, err = fb.GetInfo(r.URL, c.FileBrowser, c.User)
		if err != nil {
			return ErrorToHTTP(err, false), err
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
	case "share":
		code, err = shareHandler(c, w, r)
	case "subtitles":
		code, err = subtitlesHandler(c, w, r)
	case "subtitle":
		code, err = subtitleHandler(c, w, r)
	default:
		code = http.StatusNotFound
	}

	return code, err
}

// serveChecksum calculates the hash of a file. Supports MD5, SHA1, SHA256 and SHA512.
func checksumHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	query := r.URL.Query().Get("algo")

	val, err := c.File.Checksum(query)
	if err == fb.ErrInvalidOption {
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
func renderFile(c *fb.Context, w http.ResponseWriter, file string) (int, error) {
	tpl := template.Must(template.New("file").Parse(c.Assets.MustString(file)))

	var contentType string
	switch filepath.Ext(file) {
	case ".html":
		contentType = "text/html"
	case ".js":
		contentType = "application/javascript"
	case ".json":
		contentType = "application/json"
	default:
		contentType = "text"
	}

	w.Header().Set("Content-Type", contentType+"; charset=utf-8")

	data := map[string]interface{}{
		"baseurl":       c.RootURL(),
		"NoAuth":        c.Auth.Method == "none",
		"Version":       fb.Version,
		"CSS":           template.CSS(c.CSS),
		"ReCaptcha":     c.ReCaptcha.Key != "" && c.ReCaptcha.Secret != "",
		"ReCaptchaHost": c.ReCaptcha.Host,
		"ReCaptchaKey":  c.ReCaptcha.Key,
	}

	if c.StaticGen != nil {
		data["staticgen"] = c.StaticGen.Name()
	}

	err := tpl.Execute(w, data)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

// sharePage build the share page.
func sharePage(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	s, err := c.Store.Share.Get(r.URL.Path)
	if err == fb.ErrNotExist {
		w.WriteHeader(http.StatusNotFound)
		return renderFile(c, w, "static/share/404.html")
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	if s.Expires && s.ExpireDate.Before(time.Now()) {
		c.Store.Share.Delete(s.Hash)
		w.WriteHeader(http.StatusNotFound)
		return renderFile(c, w, "static/share/404.html")
	}

	r.URL.Path = s.Path

	info, err := os.Stat(s.Path)
	if err != nil {
		c.Store.Share.Delete(s.Hash)
		return ErrorToHTTP(err, false), err
	}

	c.File = &fb.File{
		Path:    s.Path,
		Name:    info.Name(),
		ModTime: info.ModTime(),
		Mode:    info.Mode(),
		IsDir:   info.IsDir(),
		Size:    info.Size(),
	}

	dl := r.URL.Query().Get("dl")

	if dl == "" || dl == "0" {
		tpl := template.Must(template.New("file").Parse(c.Assets.MustString("static/share/index.html")))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		err := tpl.Execute(w, map[string]interface{}{
			"baseurl": c.RootURL(),
			"File":    c.File,
		})

		if err != nil {
			return http.StatusInternalServerError, err
		}
		return 0, nil
	}

	return downloadHandler(c, w, r)
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

// ErrorToHTTP converts errors to HTTP Status Code.
func ErrorToHTTP(err error, gone bool) int {
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

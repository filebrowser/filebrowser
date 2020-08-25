package http

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	rice "github.com/GeertJohan/go.rice"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/version"
)

func handleWithStaticData(w http.ResponseWriter, _ *http.Request, d *data, box *rice.Box, file, contentType string) (int, error) {
	w.Header().Set("Content-Type", contentType)

	auther, err := d.store.Auth.Get(d.settings.AuthMethod)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	data := map[string]interface{}{
		"Name":            d.settings.Branding.Name,
		"DisableExternal": d.settings.Branding.DisableExternal,
		"BaseURL":         d.server.BaseURL,
		"Version":         version.Version,
		"StaticURL":       path.Join(d.server.BaseURL, "/static"),
		"Signup":          d.settings.Signup,
		"NoAuth":          d.settings.AuthMethod == auth.MethodNoAuth,
		"AuthMethod":      d.settings.AuthMethod,
		"LoginPage":       auther.LoginPage(),
		"CSS":             false,
		"ReCaptcha":       false,
		"Theme":           d.settings.Branding.Theme,
		"EnableThumbs":    d.server.EnableThumbnails,
		"ResizePreview":   d.server.ResizePreview,
	}

	if d.settings.Branding.Files != "" {
		fPath := filepath.Join(d.settings.Branding.Files, "custom.css")
		_, err := os.Stat(fPath) //nolint:shadow

		if err != nil && !os.IsNotExist(err) {
			log.Printf("couldn't load custom styles: %v", err)
		}

		if err == nil {
			data["CSS"] = true
		}
	}

	if d.settings.AuthMethod == auth.MethodJSONAuth {
		raw, err := d.store.Auth.Get(d.settings.AuthMethod) //nolint:shadow
		if err != nil {
			return http.StatusInternalServerError, err
		}

		auther := raw.(*auth.JSONAuth)

		if auther.ReCaptcha != nil {
			data["ReCaptcha"] = auther.ReCaptcha.Key != "" && auther.ReCaptcha.Secret != ""
			data["ReCaptchaHost"] = auther.ReCaptcha.Host
			data["ReCaptchaKey"] = auther.ReCaptcha.Key
		}
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	data["Json"] = string(b)

	index := template.Must(template.New("index").Delims("[{[", "]}]").Parse(box.MustString(file)))
	err = index.Execute(w, data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

func getStaticHandlers(store *storage.Storage, server *settings.Server) (index, static http.Handler) {
	box := rice.MustFindBox("../frontend/dist")
	handler := http.FileServer(box.HTTPBox())

	index = handle(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.Method != http.MethodGet {
			return http.StatusNotFound, nil
		}

		w.Header().Set("x-xss-protection", "1; mode=block")
		return handleWithStaticData(w, r, d, box, "index.html", "text/html; charset=utf-8")
	}, "", store, server)

	static = handle(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.Method != http.MethodGet {
			return http.StatusNotFound, nil
		}

		if d.settings.Branding.Files != "" {
			if strings.HasPrefix(r.URL.Path, "img/") {
				fPath := filepath.Join(d.settings.Branding.Files, r.URL.Path)
				if _, err := os.Stat(fPath); err == nil {
					http.ServeFile(w, r, fPath)
					return 0, nil
				}
			} else if r.URL.Path == "custom.css" && d.settings.Branding.Files != "" {
				http.ServeFile(w, r, filepath.Join(d.settings.Branding.Files, "custom.css"))
				return 0, nil
			}
		}

		if !strings.HasSuffix(r.URL.Path, ".js") {
			handler.ServeHTTP(w, r)
			return 0, nil
		}

		return handleWithStaticData(w, r, d, box, r.URL.Path, "application/javascript; charset=utf-8")
	}, "/static/", store, server)

	return index, static
}

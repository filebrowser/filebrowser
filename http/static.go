package http

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/GeertJohan/go.rice"
	"github.com/filebrowser/filebrowser/auth"
	"github.com/filebrowser/filebrowser/types"
)

func (e *Env) getStaticHandlers() (http.Handler, http.Handler) {
	box := rice.MustFindBox("../frontend/dist")
	handler := http.FileServer(box.HTTPBox())

	baseURL := strings.TrimSuffix(e.Settings.BaseURL, "/")
	staticURL := strings.TrimPrefix(baseURL+"/static", "/")
	cssFile := ""

	// TODO: baseurl must always not have the trailing slash
	data := map[string]interface{}{
		"Name":            e.Settings.Branding.Name,
		"DisableExternal": e.Settings.Branding.DisableExternal,
		"BaseURL":         baseURL,
		"Version":         types.Version,
		"StaticURL":       staticURL,
		"Signup":          e.Settings.Signup,
		"CSS":             false,
		"ReCaptcha":       false,
	}

	if e.Settings.Branding.Files != "" {
		path := filepath.Join(e.Settings.Branding.Files, "custom.css")
		_, err := os.Stat(path)

		if err != nil && !os.IsNotExist(err) {
			log.Printf("couldn't load custom styles: %v", err)
		}

		if err == nil {
			cssFile = path
			data["CSS"] = true
		}
	}

	if e.Settings.AuthMethod == auth.MethodJSONAuth {
		auther := e.Auther.(*auth.JSONAuth)

		if auther.ReCaptcha != nil {
			data["ReCaptcha"] = auther.ReCaptcha.Key != "" && auther.ReCaptcha.Secret != ""
			data["ReCaptchaHost"] = auther.ReCaptcha.Host
			data["ReCaptchaKey"] = auther.ReCaptcha.Key
		}
	}

	b, _ := json.MarshalIndent(data, "", "  ")
	data["Json"] = string(b)

	handleWithData := func(w http.ResponseWriter, r *http.Request, file string, contentType string) {
		w.Header().Set("Content-Type", contentType)
		index := template.Must(template.New("index").Delims("[{[", "]}]").Parse(box.MustString(file)))
		err := index.Execute(w, data)

		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
		}
	}

	index := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			httpErr(w, r, http.StatusNotFound, nil)
			return
		}

		w.Header().Set("x-frame-options", "SAMEORIGIN")
		w.Header().Set("x-xss-protection", "1; mode=block")

		handleWithData(w, r, "index.html", "text/html; charset=utf-8")
	})

	static := http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if e.Settings.Branding.Files != "" {
			if strings.HasPrefix(r.URL.Path, "img/") {
				path := filepath.Join(e.Settings.Branding.Files, r.URL.Path)
				if _, err := os.Stat(path); err == nil {
					http.ServeFile(w, r, path)
					return
				}
			} else if r.URL.Path == "custom.css" && cssFile != "" {
				http.ServeFile(w, r, cssFile)
				return
			}
		}

		if !strings.HasSuffix(r.URL.Path, ".js") {
			handler.ServeHTTP(w, r)
			return
		}

		handleWithData(w, r, r.URL.Path, "application/javascript; charset=utf-8")
	}))

	return index, static
}

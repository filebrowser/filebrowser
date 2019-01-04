package http

/*
import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/filebrowser/filebrowser/auth"
	"github.com/filebrowser/filebrowser/storage"
	"github.com/filebrowser/filebrowser/version"
)

func getStaticData(storage *storage.Storage) (map[string]interface{}, error) {
	settings, err := storage.Settings.Get()
	if err != nil {
		return nil, err
	}

	staticURL := strings.TrimPrefix(settings.BaseURL+"/static", "/")

	data := map[string]interface{}{
		"Name":            settings.Branding.Name,
		"DisableExternal": settings.Branding.DisableExternal,
		"BaseURL":         settings.BaseURL,
		"Version":         version.Version,
		"StaticURL":       staticURL,
		"Signup":          settings.Signup,
		"NoAuth":          settings.AuthMethod == auth.MethodNoAuth,
		"CSS":             false,
		"ReCaptcha":       false,
	}

	if settings.Branding.Files != "" {
		path := filepath.Join(settings.Branding.Files, "custom.css")
		_, err := os.Stat(path)

		if err != nil && !os.IsNotExist(err) {
			log.Printf("couldn't load custom styles: %v", err)
		}

		if err == nil {
			data["CSS"] = true
		}
	}

	if settings.AuthMethod == auth.MethodJSONAuth {
		raw, err := storage.Auth.Get()
		if err != nil {
			return nil, err
		}

		auther := raw.(*auth.JSONAuth)

		if auther.ReCaptcha != nil {
			data["ReCaptcha"] = auther.ReCaptcha.Key != "" && auther.ReCaptcha.Secret != ""
			data["ReCaptchaHost"] = auther.ReCaptcha.Host
			data["ReCaptchaKey"] = auther.ReCaptcha.Key
		}
	}

	b, _ := json.MarshalIndent(data, "", "  ")
	data["Json"] = string(b)
	return data, nil
}

func getStaticHandlers(storage *storage.Storage) (http.Handler, http.Handler, error) {
	box := rice.MustFindBox("../frontend/dist")
	handler := http.FileServer(box.HTTPBox())

	handleWithData := func(w http.ResponseWriter, r *http.Request, file string, contentType string) {
		w.Header().Set("Content-Type", contentType)
		index := template.Must(template.New("index").Delims("[{[", "]}]").Parse(box.MustString(file)))
		data, err := getStaticData(storage)
		if err != nil {

		}

		err := index.Execute(w, e.getStaticData())

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
		e.RLockSettings()
		defer e.RUnlockSettings()

		if e.GetSettings().Branding.Files != "" {
			if strings.HasPrefix(r.URL.Path, "img/") {
				path := filepath.Join(e.GetSettings().Branding.Files, r.URL.Path)
				if _, err := os.Stat(path); err == nil {
					http.ServeFile(w, r, path)
					return
				}
			} else if r.URL.Path == "custom.css" && e.GetSettings().Branding.Files != "" {
				http.ServeFile(w, r, filepath.Join(e.GetSettings().Branding.Files, "custom.css"))
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

*/

package http

import (
	"net/http"
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

	// TODO: baseurl must always not have the trailing slash
	data := map[string]interface{}{
		"BaseURL":   baseURL,
		"Version":   types.Version,
		"StaticURL": staticURL,
		"Signup":    e.Settings.Signup,
		"ReCaptcha": false,
	}

	if e.Settings.AuthMethod == auth.MethodJSONAuth {
		auther := e.Auther.(*auth.JSONAuth)

		if auther.ReCaptcha != nil {
			data["ReCaptcha"] = auther.ReCaptcha.Key != "" && auther.ReCaptcha.Secret != ""
			data["ReCaptchaHost"] = auther.ReCaptcha.Host
			data["ReCaptchaKey"] = auther.ReCaptcha.Key
		}
	}

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

	static := http.StripPrefix("/static", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ".js") {
			handler.ServeHTTP(w, r)
			return
		}

		handleWithData(w, r, strings.TrimPrefix(r.URL.Path, "/"), "application/javascript; charset=utf-8")
	}))

	return index, static
}

package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/GeertJohan/go.rice"
	"github.com/filebrowser/filebrowser/types"
	"github.com/gorilla/mux"
)

type key int

const (
	keyUserID key = iota
)

// Env ...
type Env struct {
	Auther   types.Auther
	Runner   *types.Runner
	Settings *types.Settings
	Store    *types.Store
}

func (e *Env) getHandlers() (http.Handler, http.Handler) {
	box := rice.MustFindBox("../frontend/dist")
	handler := http.FileServer(box.HTTPBox())

	baseURL := strings.TrimSuffix(e.Settings.BaseURL, "/")
	staticURL := strings.TrimPrefix(baseURL+"/static", "/")

	// TODO: baseurl must always not have the trailing slash
	data := map[string]interface{}{
		"BaseURL":   baseURL,
		"StaticURL": staticURL,
		"Signup":    e.Settings.Signup,
	}

	index := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			httpErr(w, r, http.StatusNotFound, nil)
			return
		}

		w.Header().Set("x-frame-options", "SAMEORIGIN")
		w.Header().Set("x-xss-protection", "1; mode=block")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		index := template.Must(template.New("index").Delims("[{[", "]}]").Parse(box.MustString("/index.html")))
		err := index.Execute(w, data)

		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
		}
	})

	static := http.StripPrefix("/static", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ".js") {
			handler.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		index := template.Must(template.New("index").Delims("[{[", "]}]").Parse(box.MustString(r.URL.Path)))
		err := index.Execute(w, data)

		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
		}
	}))

	return index, static
}

// Handler ...
func Handler(e *Env) http.Handler {
	r := mux.NewRouter()

	index, static := e.getHandlers()

	r.PathPrefix("/static").Handler(static)
	r.NotFoundHandler = index

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/login", e.loginHandler)
	api.HandleFunc("/signup", e.signupHandler)

	users := api.PathPrefix("/users").Subrouter()
	users.HandleFunc("", e.auth(e.usersGetHandler)).Methods("GET")
	users.HandleFunc("", e.auth(e.userPostHandler)).Methods("POST")
	users.HandleFunc("/{id:[0-9]+}", e.auth(e.userPutHandler)).Methods("PUT")
	users.HandleFunc("/{id:[0-9]+}", e.auth(e.userGetHandler)).Methods("GET")
	users.HandleFunc("/{id:[0-9]+}", e.auth(e.userDeleteHandler)).Methods("DELETE")

	api.PathPrefix("/resources").HandlerFunc(e.auth(e.resourceGetHandler)).Methods("GET")
	api.PathPrefix("/resources").HandlerFunc(e.auth(e.resourceDeleteHandler)).Methods("DELETE")
	api.PathPrefix("/resources").HandlerFunc(e.auth(e.resourcePostPutHandler)).Methods("POST")
	api.PathPrefix("/resources").HandlerFunc(e.auth(e.resourcePostPutHandler)).Methods("PUT")
	api.PathPrefix("/resources").HandlerFunc(e.auth(e.resourcePatchHandler)).Methods("PATCH")

	api.PathPrefix("/share").HandlerFunc(e.auth(e.shareGetHandler)).Methods("GET")
	api.PathPrefix("/share").HandlerFunc(e.auth(e.sharePostHandler)).Methods("POST")
	api.PathPrefix("/share").HandlerFunc(e.auth(e.shareDeleteHandler)).Methods("DELETE")

	api.PathPrefix("/raw").HandlerFunc(e.auth(e.rawHandler)).Methods("GET")

	return r
}

func httpErr(w http.ResponseWriter, r *http.Request, status int, err error) {
	txt := http.StatusText(status)
	if err != nil || status >= 400 {
		log.Printf("%s: %v %s %v", r.URL.Path, status, r.RemoteAddr, err)
	}
	http.Error(w, strconv.Itoa(status)+" "+txt, status)
}

func renderJSON(w http.ResponseWriter, r *http.Request, data interface{}) {
	marsh, err := json.Marshal(data)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(marsh); err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
	}
}

func (e *Env) getUser(w http.ResponseWriter, r *http.Request) (*types.User, bool) {
	id := r.Context().Value(keyUserID).(uint)
	user, err := e.Store.Users.Get(id)
	if err == types.ErrNotExist {
		httpErr(w, r, http.StatusForbidden, nil)
		return nil, false
	}

	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return nil, false
	}

	return user, true
}

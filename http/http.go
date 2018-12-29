package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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

// Handler ...
func Handler(e *Env) http.Handler {
	r := mux.NewRouter()

	index, static := e.getStaticHandlers()

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

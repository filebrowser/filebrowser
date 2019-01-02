package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/filebrowser/filebrowser/types"
	"github.com/gorilla/mux"
)

type key int

const (
	keyUserID key = iota
)

type modifyRequest struct {
	What  string   `json:"what"`  // Answer to: what data type?
	Which []string `json:"which"` // Answer to: which fields?
}

// Env contains the required info for FB to run.
type Env struct {
	Auther   types.Auther
	Settings *types.Settings
	Store    *types.Store
	mux      sync.Mutex // settings mutex for Auther, Runner and Settings changes.
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
	api.HandleFunc("/renew", e.auth(e.renew))

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

	api.HandleFunc("/settings", e.auth(e.settingsGetHandler)).Methods("GET")
	api.HandleFunc("/settings", e.auth(e.settingsPutHandler)).Methods("PUT")

	api.PathPrefix("/raw").HandlerFunc(e.auth(e.rawHandler)).Methods("GET")
	api.PathPrefix("/command").HandlerFunc(e.auth(e.commandsHandler))
	api.PathPrefix("/search").HandlerFunc(e.auth(e.searchHandler))

	return r
}

func httpErr(w http.ResponseWriter, r *http.Request, status int, err error) {
	txt := http.StatusText(status)
	if err != nil || status >= 400 {
		log.Printf("%s: %v %s %v", r.URL.Path, status, r.RemoteAddr, err)
	}
	http.Error(w, strconv.Itoa(status)+" "+txt, status)
}

func wsErr(ws *websocket.Conn, r *http.Request, status int, err error) {
	txt := http.StatusText(status)
	if err != nil || status >= 400 {
		log.Printf("%s: %v %s %v", r.URL.Path, status, r.RemoteAddr, err)
	}
	ws.WriteControl(websocket.CloseInternalServerErr, []byte(txt), time.Now().Add(10*time.Second))
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

func (e *Env) getAdminUser(w http.ResponseWriter, r *http.Request) (*types.User, bool) {
	user, ok := e.getUser(w, r)
	if !ok {
		return nil, false
	}

	if !user.Perm.Admin {
		httpErr(w, r, http.StatusForbidden, nil)
		return nil, false
	}

	return user, true
}

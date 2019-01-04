package http

import (
	"net/http"

	"github.com/filebrowser/filebrowser/storage"
	"github.com/gorilla/mux"
)

type modifyRequest struct {
	What  string   `json:"what"`  // Answer to: what data type?
	Which []string `json:"which"` // Answer to: which fields?
}

func NewHandler(storage *storage.Storage) (http.Handler, error) {
	r := mux.NewRouter()

	/* index, static := e.getStaticHandlers()

	r.PathPrefix("/static").Handler(static)
	r.NotFoundHandler = index */

	api := r.PathPrefix("/api").Subrouter()

	api.Handle("/login", handle(loginHandler, "", storage))
	api.Handle("/signup", handle(signupHandler, "", storage))
	api.Handle("/renew", handle(renewHandler, "", storage))

	/*  users := api.PathPrefix("/users").Subrouter()
	users.HandleFunc("", e.auth(e.usersGetHandler)).Methods("GET")
	users.HandleFunc("", e.auth(e.userPostHandler)).Methods("POST")
	users.HandleFunc("/{id:[0-9]+}", e.auth(e.userPutHandler)).Methods("PUT")
	users.HandleFunc("/{id:[0-9]+}", e.auth(e.userGetHandler)).Methods("GET")
	users.HandleFunc("/{id:[0-9]+}", e.auth(e.userDeleteHandler)).Methods("DELETE") */

	api.PathPrefix("/resources").Handler(handle(resourceGetHandler, "/api/resources", storage)).Methods("GET")
	api.PathPrefix("/resources").Handler(handle(resourceDeleteHandler, "/api/resources", storage)).Methods("DELETE")
	api.PathPrefix("/resources").Handler(handle(resourcePostPutHandler, "/api/resources", storage)).Methods("POST")
	api.PathPrefix("/resources").Handler(handle(resourcePostPutHandler, "/api/resources", storage)).Methods("PUT")
	api.PathPrefix("/resources").Handler(handle(resourcePatchHandler, "/api/resources", storage)).Methods("PATCH")

	api.PathPrefix("/share").Handler(handle(shareGetHandler, "/api/share", storage)).Methods("GET")
	api.PathPrefix("/share").Handler(handle(sharePostHandler, "/api/share", storage)).Methods("POST")
	api.PathPrefix("/share").Handler(handle(shareDeleteHandler, "/api/share", storage)).Methods("DELETE")

	api.Handle("/settings", handle(settingsGetHandler, "", storage)).Methods("GET")
	api.Handle("/settings", handle(settingsPutHandler, "", storage)).Methods("PUT")

	/* api.PathPrefix("/raw").HandlerFunc(e.auth(e.rawHandler)).Methods("GET")
	api.PathPrefix("/command").HandlerFunc(e.auth(e.commandsHandler))
	api.PathPrefix("/search").HandlerFunc(e.auth(e.searchHandler)) */

	return r, nil
}


/*
type key int





func wsErr(ws *websocket.Conn, r *http.Request, status int, err error) {
	txt := http.StatusText(status)
	if err != nil || status >= 400 {
		log.Printf("%s: %v %s %v", r.URL.Path, status, r.RemoteAddr, err)
	}
	ws.WriteControl(websocket.CloseInternalServerErr, []byte(txt), time.Now().Add(10*time.Second))
}
*/

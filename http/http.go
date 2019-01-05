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

	index, static := getStaticHandlers(storage)

	r.PathPrefix("/static").Handler(static)
	r.NotFoundHandler = index

	api := r.PathPrefix("/api").Subrouter()

	api.Handle("/login", handle(loginHandler, "", storage))
	api.Handle("/signup", handle(signupHandler, "", storage))
	api.Handle("/renew", handle(renewHandler, "", storage))

	users := api.PathPrefix("/users").Subrouter()
	users.Handle("", handle(usersGetHandler, "", storage)).Methods("GET")
	users.Handle("", handle(userPostHandler, "", storage)).Methods("POST")
	users.Handle("/{id:[0-9]+}", handle(userPutHandler, "", storage)).Methods("PUT")
	users.Handle("/{id:[0-9]+}", handle(userGetHandler, "", storage)).Methods("GET")
	users.Handle("/{id:[0-9]+}", handle(userDeleteHandler, "", storage)).Methods("DELETE")

	api.PathPrefix("/resources").Handler(handle(resourceGetHandler, "/api/resources", storage)).Methods("GET")
	api.PathPrefix("/resources").Handler(handle(resourceDeleteHandler, "/api/resources", storage)).Methods("DELETE")
	api.PathPrefix("/resources").Handler(handle(resourcePostPutHandler, "/api/resources", storage)).Methods("POST")
	api.PathPrefix("/resources").Handler(handle(resourcePostPutHandler, "/api/resources", storage)).Methods("PUT")
	api.PathPrefix("/resources").Handler(handle(resourcePatchHandler, "/api/resources", storage)).Methods("PATCH")

	api.PathPrefix("/share").Handler(handle(shareGetsHandler, "/api/share", storage)).Methods("GET")
	api.PathPrefix("/share").Handler(handle(sharePostHandler, "/api/share", storage)).Methods("POST")
	api.PathPrefix("/share").Handler(handle(shareDeleteHandler, "/api/share", storage)).Methods("DELETE")

	api.Handle("/settings", handle(settingsGetHandler, "", storage)).Methods("GET")
	api.Handle("/settings", handle(settingsPutHandler, "", storage)).Methods("PUT")

	api.PathPrefix("/raw").Handler(handle(rawHandler, "/api/raw", storage)).Methods("GET")
	api.PathPrefix("/command").Handler(handle(commandsHandler, "/api/command", storage)).Methods("GET")
	api.PathPrefix("/search").Handler(handle(searchHandler, "/api/search", storage)).Methods("GET")
	
	public := api.PathPrefix("/public").Subrouter()
	public.PathPrefix("/dl").Handler(handle(publicDlHandler, "/api/public/dl/", storage)).Methods("GET")
	public.PathPrefix("/share").Handler(handle(publicShareHandler, "/api/public/share/", storage)).Methods("GET")

	return r, nil
}

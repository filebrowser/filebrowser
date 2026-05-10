package fbhttp

import (
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/filebrowser/filebrowser/v2/cnc"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
)

type modifyRequest struct {
	What            string   `json:"what"`             // Answer to: what data type?
	Which           []string `json:"which"`            // Answer to: which fields?
	CurrentPassword string   `json:"current_password"` // Answer to: user logged password
}

func NewHandler(
	imgSvc ImgService,
	fileCache FileCache,
	uploadCache UploadCache,
	store *storage.Storage,
	server *settings.Server,
	assetsFs fs.FS,
	registry *cnc.Registry,
) (http.Handler, error) {
	server.Clean()

	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", `default-src 'self'; style-src 'unsafe-inline';`)
			next.ServeHTTP(w, r)
		})
	})
	index, static := getStaticHandlers(store, server, assetsFs)

	monkey := func(fn handleFunc, prefix string) http.Handler {
		return handle(fn, prefix, store, server)
	}

	r.HandleFunc("/health", healthHandler)
	r.PathPrefix("/static").Handler(static)
	r.NotFoundHandler = index

	api := r.PathPrefix("/api").Subrouter()

	tokenExpirationTime := server.GetTokenExpirationTime(DefaultTokenExpirationTime)
	api.Handle("/login", monkey(loginHandler(tokenExpirationTime), ""))
	api.Handle("/signup", monkey(signupHandler, ""))
	api.Handle("/renew", monkey(renewHandler(tokenExpirationTime), ""))

	users := api.PathPrefix("/users").Subrouter()
	users.Handle("", monkey(usersGetHandler, "")).Methods("GET")
	users.Handle("", monkey(userPostHandler, "")).Methods("POST")
	users.Handle("/{id:[0-9]+}", monkey(userPutHandler, "")).Methods("PUT")
	users.Handle("/{id:[0-9]+}", monkey(userGetHandler, "")).Methods("GET")
	users.Handle("/{id:[0-9]+}", monkey(userDeleteHandler, "")).Methods("DELETE")

	api.PathPrefix("/resources").Handler(monkey(resourceGetHandler, "/api/resources")).Methods("GET")
	api.PathPrefix("/resources").Handler(monkey(resourceDeleteHandler(fileCache), "/api/resources")).Methods("DELETE")
	api.PathPrefix("/resources").Handler(monkey(resourcePostHandler(fileCache), "/api/resources")).Methods("POST")
	api.PathPrefix("/resources").Handler(monkey(resourcePutHandler, "/api/resources")).Methods("PUT")
	api.PathPrefix("/resources").Handler(monkey(resourcePatchHandler(fileCache), "/api/resources")).Methods("PATCH")

	api.PathPrefix("/tus").Handler(monkey(tusPostHandler(uploadCache), "/api/tus")).Methods("POST")
	api.PathPrefix("/tus").Handler(monkey(tusHeadHandler(uploadCache), "/api/tus")).Methods("HEAD", "GET")
	api.PathPrefix("/tus").Handler(monkey(tusPatchHandler(uploadCache), "/api/tus")).Methods("PATCH")
	api.PathPrefix("/tus").Handler(monkey(tusDeleteHandler(uploadCache), "/api/tus")).Methods("DELETE")

	api.PathPrefix("/usage").Handler(monkey(diskUsage, "/api/usage")).Methods("GET")

	api.Handle("/shares", monkey(shareListHandler, "")).Methods("GET")
	api.PathPrefix("/share").Handler(monkey(shareGetsHandler, "/api/share")).Methods("GET")
	api.PathPrefix("/share").Handler(monkey(sharePostHandler, "/api/share")).Methods("POST")
	api.PathPrefix("/share").Handler(monkey(shareDeleteHandler, "/api/share")).Methods("DELETE")

	api.Handle("/settings", monkey(settingsGetHandler, "")).Methods("GET")
	api.Handle("/settings", monkey(settingsPutHandler, "")).Methods("PUT")

	cncRouter := api.PathPrefix("/cnc").Subrouter()
	cncRouter.Handle("/settings", monkey(cncSettingsGetHandler, "")).Methods("GET")
	cncRouter.Handle("/settings", monkey(cncSettingsPutHandler(registry), "")).Methods("PUT")
	cncRouter.Handle("/settings/token", monkey(cncRegenerateTokenHandler, "")).Methods("POST")
	cncRouter.Handle("/machines", monkey(cncMachinesListHandler(registry), "")).Methods("GET")
	cncRouter.Handle("/status", monkey(cncStatusHandler(registry), "")).Methods("GET")
	cncRouter.Handle("/check", monkey(cncCheckHandler(registry), "")).Methods("POST")
	cncRouter.Handle("/probe-tools", monkey(cncProbeToolsHandler(registry), "")).Methods("POST")
	cncRouter.Handle("/probe-tool-life", monkey(cncProbeToolLifeHandler(registry), "")).Methods("POST")
	cncRouter.Handle("/tool-table", monkey(cncToolTableReadHandler(registry), "")).Methods("POST")
	cncRouter.Handle("/tool-table", monkey(cncToolTableLatestHandler(registry), "")).Methods("GET")
	cncRouter.Handle("/tool-table/history", monkey(cncToolTableHistoryHandler(registry), "")).Methods("GET")
	cncRouter.Handle("/preflight", monkey(cncPreflightHandler(registry), "")).Methods("GET")
	cncRouter.Handle("/siblings", monkey(cncSiblingsHandler(registry), "")).Methods("GET")
	cncRouter.Handle("/start", monkey(cncStartHandler(registry), "")).Methods("POST")
	cncRouter.Handle("/stop", monkey(cncStopHandler(registry), "")).Methods("POST")
	cncRouter.Handle("/qcode", monkey(cncQueryHandler(registry), "")).Methods("POST")
	cncRouter.Handle("/stream", monkey(cncStreamHandler(registry), "")).Methods("GET")
	cncRouter.Handle("/state", monkey(cncStateHandler(registry), "")).Methods("GET")
	cncRouter.Handle("/recovery/ack", monkey(cncRecoveryAckHandler(registry), "")).Methods("POST")

	api.PathPrefix("/raw").Handler(monkey(rawHandler, "/api/raw")).Methods("GET")
	api.PathPrefix("/preview/{size}/{path:.*}").
		Handler(monkey(previewHandler(imgSvc, fileCache, server.EnableThumbnails, server.ResizePreview), "/api/preview")).Methods("GET")
	api.PathPrefix("/command").Handler(monkey(commandsHandler, "/api/command")).Methods("GET")
	api.PathPrefix("/search").Handler(monkey(searchHandler, "/api/search")).Methods("GET")
	api.PathPrefix("/subtitle").Handler(monkey(subtitleHandler, "/api/subtitle")).Methods("GET")

	public := api.PathPrefix("/public").Subrouter()
	public.PathPrefix("/dl").Handler(monkey(publicDlHandler, "/api/public/dl/")).Methods("GET")
	public.PathPrefix("/share").Handler(monkey(publicShareHandler, "/api/public/share/")).Methods("GET")

	return stripPrefix(server.BaseURL, r), nil
}

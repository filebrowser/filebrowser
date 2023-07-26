package http

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	tusd "github.com/tus/tusd/pkg/handler"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
)

type TusHandler struct {
	store        *storage.Storage
	server       *settings.Server
	settings     *settings.Settings
	tusdHandlers map[uint]*tusd.UnroutedHandler
	apiPath      string
	mutex        *sync.Mutex
}

func NewTusHandler(store *storage.Storage, server *settings.Server, apiPath string) (_ *TusHandler, err error) {
	tusHandler := &TusHandler{}
	tusHandler.store = store
	tusHandler.server = server
	tusHandler.tusdHandlers = make(map[uint]*tusd.UnroutedHandler)
	tusHandler.apiPath = apiPath
	tusHandler.mutex = &sync.Mutex{}

	if tusHandler.settings, err = store.Settings.Get(); err != nil {
		return tusHandler, fmt.Errorf("couldn't get settings: %w", err)
	}

	return tusHandler, nil
}

func (th *TusHandler) getOrCreateTusdHandler(d *data, r *http.Request) (_ *tusd.UnroutedHandler, err error) {
	// Use a mutex to make sure only one tus handler is created for each user
	th.mutex.Lock()
	defer th.mutex.Unlock()

	tusdHandler, ok := th.tusdHandlers[d.user.ID]
	if !ok {
		// If we don't define an absolute URL for tusd, it creates an absolute URL for us that the client will use.
		// See tusd/handler/unrouted_handler.go/absFileURL() for details.
		// This URL's scheme will be http in our case (as we don't use tusd's inbuilt TLS feature),
		// which is fine if we don't use both a browser and a reverse proxy that terminates SSL for us.
		// In case we do, we need to define an absolute URL with the correct scheme, or we'll get mixed content errors.
		// We can extract the correct scheme and host from the origin request header, if it exists (which always is the case for browsers).
		var origin string
		if originHeader, ok := r.Header["Origin"]; ok && len(originHeader) > 0 {
			origin = originHeader[0]
		}
		basePath, err := url.JoinPath(origin, th.server.BaseURL, th.apiPath)
		if err != nil {
			return nil, err
		}

		log.Printf("Creating tus handler for user %s on path %s\n", d.user.Username, basePath)
		tusdHandler, err = th.createTusdHandler(d, basePath)
		if err != nil {
			return nil, err
		}
		th.tusdHandlers[d.user.ID] = tusdHandler
	}

	return tusdHandler, nil
}

func (th TusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, err := withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		// Check if user has permission to create files
		if !d.user.Perm.Create {
			return http.StatusForbidden, nil
		}

		// Create a new tus handler for current user if it doesn't exist yet
		tusdHandler, err := th.getOrCreateTusdHandler(d, r)
		if err != nil {
			return http.StatusBadRequest, err
		}

		switch r.Method {
		case "POST":
			tusdHandler.PostFile(w, r)
		case "HEAD":
			tusdHandler.HeadFile(w, r)
		case "PATCH":
			tusdHandler.PatchFile(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

		// Isn't used
		return http.StatusNoContent, nil
	})(w, r, &data{
		store:    th.store,
		settings: th.settings,
		server:   th.server,
	})

	switch {
	case err != nil:
		http.Error(w, err.Error(), code)
	case code >= http.StatusBadRequest:
		http.Error(w, "", code)
	}
}

func (th TusHandler) createTusdHandler(d *data, basePath string) (*tusd.UnroutedHandler, error) {
	tusStore := NewInPlaceDataStore(d.user.FullPath("/"), d.user.Perm.Modify)
	composer := tusd.NewStoreComposer()
	tusStore.UseIn(composer)

	tusdHandler, err := tusd.NewUnroutedHandler(tusd.Config{
		BasePath:      basePath,
		StoreComposer: composer,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create tusdHandler: %w", err)
	}

	return tusdHandler, nil
}

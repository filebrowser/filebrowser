package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"

	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
)

const uploadDirName = ".tmp_upload"

type TusHandler struct {
	store                *storage.Storage
	server               *settings.Server
	settings             *settings.Settings
	tusdHandlers         map[uint]*tusd.UnroutedHandler
	notifyNewTusdHandler chan struct{}
	apiPath              string
	mutex                *sync.Mutex
}

func NewTusHandler(store *storage.Storage, server *settings.Server, apiPath string) (TusHandler, error) {
	tusHandler := TusHandler{}
	tusHandler.store = store
	tusHandler.server = server
	tusHandler.tusdHandlers = make(map[uint]*tusd.UnroutedHandler)
	tusHandler.notifyNewTusdHandler = make(chan struct{})
	tusHandler.apiPath = apiPath
	tusHandler.mutex = &sync.Mutex{}

	var err error
	if tusHandler.settings, err = store.Settings.Get(); err != nil {
		return tusHandler, fmt.Errorf("couldn't get settings: %w", err)
	}

	// Create a goroutine that handles uploaded file events for all users
	go tusHandler.handleFileUploadedEvents()

	return tusHandler, nil
}

func (th TusHandler) getOrCreateTusdHandler(d *data, r *http.Request) (*tusd.UnroutedHandler, error) {
	// Use a mutex to make sure only one tus handler is created for each user
	th.mutex.Lock()
	defer th.mutex.Unlock()

	tusdHandler, ok := th.tusdHandlers[d.user.ID]
	log.Printf("Getting tus handler for user %s with basePath %s\n", d.user.Username, d.user.FullPath("/"))
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
		tusdHandler, err := th.createTusdHandler(d, basePath) //nolint:govet
		if err != nil {
			return nil, err
		}
		th.tusdHandlers[d.user.ID] = tusdHandler
		th.notifyNewTusdHandler <- struct{}{}
	}

	return tusdHandler, nil
}

func (th TusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, err := withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		// Create a new tus handler for current user if it doesn't exist yet
		tusdHandler, err := th.getOrCreateTusdHandler(d, r)

		if err != nil {
			return http.StatusBadRequest, err
		}

		// Create upload directory for each request
		uploadDir := filepath.Join(d.user.FullPath("/"), uploadDirName)
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			return http.StatusInternalServerError, err
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
	uploadDir := filepath.Join(d.user.FullPath("/"), uploadDirName)
	tusStore := filestore.FileStore{
		Path: uploadDir,
	}
	composer := tusd.NewStoreComposer()
	tusStore.UseIn(composer)

	tusdHandler, err := tusd.NewUnroutedHandler(tusd.Config{
		BasePath:              basePath,
		StoreComposer:         composer,
		NotifyCompleteUploads: true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create tusdHandler: %w", err)
	}

	return tusdHandler, nil
}

func getMetadataField(metadata tusd.MetaData, field string) (string, error) {
	if value, ok := metadata[field]; ok {
		return value, nil
	} else {
		return "", fmt.Errorf("metadata field %s not found in upload request", field)
	}
}

func (th TusHandler) handleFileUploadedEvents() {
	// Instead of running a goroutine for each user, we use a single goroutine that handles events for all users.
	// This works by using a reflect select statement that waits for events from all users.
	// On top of this, the reflect select statement also waits for a notification channel that is used to notify
	// the goroutine when a new user has been added to so that the reflect select statement can be updated.
	for {
		cases := make([]reflect.SelectCase, len(th.tusdHandlers)+1)
		// UserIDs != position in select statement, so store mapping
		caseIdsToUserIds := make(map[int]uint, len(th.tusdHandlers))
		cases[0] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(th.notifyNewTusdHandler)}
		i := 1
		for userID, tusdHandler := range th.tusdHandlers {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(tusdHandler.CompleteUploads)}
			caseIdsToUserIds[i] = userID
			i++
		}

		for {
			chosen, value, _ := reflect.Select(cases)
			if chosen == 0 {
				// Notification channel has been triggered,
				// so we need to update the reflect select statement
				break
			}

			// Get user ID from reflect select statement
			userID := caseIdsToUserIds[chosen]
			user, err := th.store.Users.Get(th.server.Root, userID)
			if err != nil {
				log.Printf("ERROR: couldn't get user with ID %d: %s\n", userID, err)
				continue
			}
			event := value.Interface().(tusd.HookEvent)
			if err := th.handleFileUploaded(user, &event); err != nil {
				log.Printf("ERROR: couldn't handle completed upload: %s\n", err)
			}
		}
	}
}

func (th TusHandler) handleFileUploaded(user *users.User, event *tusd.HookEvent) error {
	// Clean up only if an upload has been finalized
	if !event.Upload.IsFinal {
		return nil
	}

	filename, err := getMetadataField(event.Upload.MetaData, "filename")
	if err != nil {
		return err
	}
	destination, err := getMetadataField(event.Upload.MetaData, "destination")
	if err != nil {
		return err
	}
	overwriteStr, err := getMetadataField(event.Upload.MetaData, "overwrite")
	if err != nil {
		return err
	}
	userPath := user.FullPath("/")
	uploadDir := filepath.Join(userPath, uploadDirName)
	uploadedFile := filepath.Join(uploadDir, event.Upload.ID)
	fullDestination := filepath.Join(userPath, destination)

	log.Printf("Upload of %s (%s) is finished. Moving file to destination (%s) "+
		"and cleaning up temporary files.\n", filename, uploadedFile, fullDestination)

	// Check if destination file already exists. If so, we require overwrite to be set
	if _, err := os.Stat(fullDestination); !errors.Is(err, os.ErrNotExist) {
		if overwrite, err := strconv.ParseBool(overwriteStr); err != nil {
			return err
		} else if !overwrite {
			return fmt.Errorf("overwrite is set to false while destination file %s exists", destination)
		}
	}

	// Move uploaded file from tmp upload folder to user folder
	if err := os.Rename(uploadedFile, fullDestination); err != nil {
		return err
	}

	return th.removeTemporaryFiles(uploadDir, &event.Upload)
}

func (th TusHandler) removeTemporaryFiles(uploadDir string, upload *tusd.FileInfo) error {
	// Remove uploaded tmp files for finished upload (.info objects are created and need to be removed, too))
	for _, partialUpload := range append(upload.PartialUploads, upload.ID) {
		filesToDelete, err := filepath.Glob(filepath.Join(uploadDir, partialUpload+"*"))
		if err != nil {
			return err
		}
		for _, f := range filesToDelete {
			if err := os.Remove(f); err != nil {
				return err
			}
		}
	}

	// Delete folder basePath if it is empty after the request
	dir, err := os.ReadDir(uploadDir)
	if err != nil {
		return err
	}

	if len(dir) == 0 {
		// os.Remove won't remove non-empty folders in case of race condition
		if err := os.Remove(uploadDir); err != nil {
			return err
		}
	}

	return nil
}

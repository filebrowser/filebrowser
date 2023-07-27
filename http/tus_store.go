// InPlaceDataStore is a storage backend for tusd, which stores the uploaded
// files in the user's root directory, without creating any auxiliary files.
// It thus requires no clean-up on failed uploads.
// The destination metadata field needs to be set in the upload request.
// For each NewUpload, the target file is expanded by the upload's size.
// This way, multiple uploads can work on the same file, without interfering
// with each other.
// The uploads are resumable. Also, parallel uploads are supported, however,
// the initial POST requests to NewUpload must be synchronized and in order.
// Otherwise, no guarantee of the upload's integrity can be given.

package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	tusd "github.com/tus/tusd/pkg/handler"
)

const uidLength = 16
const filePerm = 0644

type InPlaceDataStore struct {
	// All uploads will be stored relative to this directory.
	// It equals the user's root directory.
	path string

	// Store whether the user is permitted to create new files.
	createPerm bool

	// Store whether the user is permitted to modify files or only create new ones.
	modifyPerm bool

	// Maps an upload ID to its object.
	// Required, since GetUpload only provides us with the id of an upload
	// and expects us to return the Info object.
	uploadsByID map[string]*InPlaceUpload

	// Map all uploads by their path.
	// Each path can have multiple uploads, as multiple uploads can work on the same file
	// when parallel uploads are enabled.
	uploadsByPath map[string][]*InPlaceUpload

	// Each upload appends to the file, so we need to make sure
	// each upload has expanded the file by info.Size bytes, before the next
	// upload is created.
	mutex *sync.Mutex
}

func NewInPlaceDataStore(path string, createPerm, modifyPerm bool) *InPlaceDataStore {
	return &InPlaceDataStore{
		path:          path,
		createPerm:    createPerm,
		modifyPerm:    modifyPerm,
		uploadsByID:   make(map[string]*InPlaceUpload),
		uploadsByPath: make(map[string][]*InPlaceUpload),
		mutex:         &sync.Mutex{},
	}
}

func (store *InPlaceDataStore) UseIn(composer *tusd.StoreComposer) {
	composer.UseCore(store)
	composer.UseConcater(store)
}

func (store *InPlaceDataStore) isPartOfNewUpload(fileExists bool, filePath string) bool {
	if !fileExists {
		// If the file doesn't exist, remove all upload references.
		// This way we can eliminate inconsistencies for failed uploads.
		for _, upload := range store.uploadsByPath[filePath] {
			delete(store.uploadsByID, upload.ID)
		}
		delete(store.uploadsByPath, filePath)

		return true
	}

	// In case the file exists, it is still possible that it is a new upload.
	// E.g.: the user wants to overwrite an existing file.
	return store.uploadsByPath[filePath] == nil
}

func (store *InPlaceDataStore) checkPermissions(isPartOfNewUpload bool) error {
	// Return tusd.HTTPErrors, as they are handled by tusd.
	if isPartOfNewUpload {
		if !store.createPerm {
			return tusd.NewHTTPError(errors.New("user is not allowed to create a new upload"), http.StatusForbidden)
		}
	}

	if !store.modifyPerm {
		return tusd.NewHTTPError(errors.New("user is not allowed to modify existing files"), http.StatusForbidden)
	}

	return nil
}

func (store *InPlaceDataStore) initializeUpload(filePath string, info *tusd.FileInfo) (int64, error) {
	fileExists := true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fileExists = false
	} else if err != nil {
		return 0, err
	}

	// Delete existing files and references, if necessary.
	isPartOfNewUpload := store.isPartOfNewUpload(fileExists, filePath)

	if err := store.checkPermissions(isPartOfNewUpload); err != nil {
		return 0, err
	}
	if isPartOfNewUpload && fileExists {
		// Remove the file's contents (instead of re-creating it).
		return 0, os.Truncate(filePath, 0)
	}
	if isPartOfNewUpload && !fileExists {
		// Create the file, if it doesn't exist.
		if _, err := os.Create(filePath); err != nil {
			return 0, err
		}
		return 0, nil
	}

	// The file exists and is part of an existing upload.
	// Open the file and enlarge it by the upload's size.
	file, err := os.OpenFile(filePath, os.O_WRONLY, filePerm)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Get the file's current size and offset to the end of the file.
	actualOffset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	// Enlarge the file by the upload's size (starting from the current offset).
	if _, err = file.Write(make([]byte, info.Size)); err != nil {
		return 0, err
	}

	return actualOffset, nil
}

func (store *InPlaceDataStore) NewUpload(ctx context.Context, info tusd.FileInfo) (_ tusd.Upload, err error) { //nolint: gocritic
	// The method must return an unique id which is used to identify the upload
	if info.ID, err = uid(); err != nil {
		return nil, err
	}

	destination, ok := info.MetaData["destination"]
	if !ok {
		return nil, errors.New("metadata field 'destination' not found in upload request")
	}
	filePath := filepath.Join(store.path, destination)

	upload := &InPlaceUpload{
		FileInfo:     info,
		filePath:     filePath,
		actualOffset: info.Size,
		parent:       store,
	}

	// Lock the mutex, as we need to modify the target file synchronously.
	store.mutex.Lock()
	defer store.mutex.Unlock()

	// Tus creates a POST request for the final concatenation.
	// In that case, we don't need to create a new upload.
	if !info.IsFinal {
		if upload.actualOffset, err = store.initializeUpload(filePath, &info); err != nil {
			return nil, err
		}
	}
	store.uploadsByID[upload.ID] = upload
	store.uploadsByPath[upload.filePath] = append(store.uploadsByPath[upload.filePath], upload)

	return upload, nil
}

func (store *InPlaceDataStore) GetUpload(ctx context.Context, id string) (tusd.Upload, error) {
	if upload, ok := store.uploadsByID[id]; ok {
		return upload, nil
	} else {
		return nil, errors.New("upload not found")
	}
}

// We need to define a concater, as client libraries will automatically ask for a concatenation.
func (store *InPlaceDataStore) AsConcatableUpload(upload tusd.Upload) tusd.ConcatableUpload {
	return upload.(*InPlaceUpload)
}

type InPlaceUpload struct {
	tusd.FileInfo
	// Extend the tusd.FileInfo struct with the target path of our uploaded file.
	filePath string
	// tusd expects offset to equal the upload's written bytes.
	// As we can have multiple uploads working on the same file,
	// this is not the case for us. Thus, store the actual offset.
	// See: https://github.com/tus/tusd/blob/main/pkg/handler/unrouted_handler.go#L714
	actualOffset int64
	// Enable the upload to remove itself from the active uploads map.
	parent *InPlaceDataStore
}

func (upload *InPlaceUpload) WriteChunk(ctx context.Context, offset int64, src io.Reader) (int64, error) {
	// Open the file and seek to the given offset.
	// Then, copy the given reader to the file, update the offset and return.
	file, err := os.OpenFile(upload.filePath, os.O_WRONLY, filePerm)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	if _, err = file.Seek(upload.actualOffset+offset, io.SeekStart); err != nil {
		return 0, err
	}

	n, err := io.Copy(file, src)
	if err != nil {
		return 0, err
	}

	upload.Offset += n
	return n, nil
}

func (upload *InPlaceUpload) GetInfo(ctx context.Context) (tusd.FileInfo, error) {
	return upload.FileInfo, nil
}

func (upload *InPlaceUpload) GetReader(ctx context.Context) (io.Reader, error) {
	return os.Open(upload.filePath)
}

func (upload *InPlaceUpload) FinishUpload(ctx context.Context) error {
	upload.parent.mutex.Lock()
	defer upload.parent.mutex.Unlock()

	delete(upload.parent.uploadsByID, upload.ID)
	uploadsByPath := upload.parent.uploadsByPath[upload.filePath]
	for i, u := range uploadsByPath {
		if u.ID == upload.ID {
			upload.parent.uploadsByPath[upload.filePath] = append(uploadsByPath[:i], uploadsByPath[i+1:]...)
			break
		}
	}
	if len(upload.parent.uploadsByPath[upload.filePath]) == 0 {
		delete(upload.parent.uploadsByPath, upload.filePath)
	}

	return nil
}

func (upload *InPlaceUpload) ConcatUploads(ctx context.Context, uploads []tusd.Upload) (err error) {
	for _, u := range uploads {
		if err := (u.(*InPlaceUpload)).FinishUpload(ctx); err != nil {
			return err
		}
	}
	return upload.FinishUpload(ctx)
}

func uid() (string, error) {
	id := make([]byte, uidLength)
	if _, err := io.ReadFull(rand.Reader, id); err != nil {
		return "", err
	}
	return hex.EncodeToString(id), nil
}

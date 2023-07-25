package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
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
	basePath string

	// Maps an upload ID to its object.
	// Required, since GetUpload only provides us with the id of an upload
	// and expects us to return the Info object.
	uploads map[string]*InPlaceUpload

	// Each upload appends to the file, so we need to make sure
	// each upload has expanded the file by info.Size bytes, before the next
	// upload is created.
	mutex *sync.Mutex
}

func NewInPlaceDataStore(basePath string) *InPlaceDataStore {
	return &InPlaceDataStore{
		basePath: basePath,
		uploads:  make(map[string]*InPlaceUpload),
		mutex:    &sync.Mutex{},
	}
}

func (store *InPlaceDataStore) UseIn(composer *tusd.StoreComposer) {
	composer.UseCore(store)
	composer.UseConcater(store)
}

func prepareFile(filePath string, uploadSize int64, mutex *sync.Mutex) (int64, error) {
	mutex.Lock()
	defer mutex.Unlock()

	// Create the file if it doesn't exist.
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, filePerm)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Get the file's current size.
	actualOffset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	// Enlarge the file by the upload's size.
	_, err = file.Write(make([]byte, uploadSize))
	if err != nil {
		return 0, err
	}

	return actualOffset, nil
}

func (store *InPlaceDataStore) NewUpload(ctx context.Context, info tusd.FileInfo) (_ tusd.Upload, err error) { //nolint: gocritic
	// The method must return an unique id which is used to identify the upload
	if info.ID == "" {
		info.ID, err = uid()
		if err != nil {
			return nil, err
		}
	}

	destination, ok := info.MetaData["destination"]
	if !ok {
		return nil, fmt.Errorf("metadata field 'destination' not found in upload request")
	}
	filePath := filepath.Join(store.basePath, destination)

	// Tus creates a POST request for the final concatenation.
	// In that case, we don't need to create a new upload.
	actualOffset := info.Size
	if info.IsPartial {
		actualOffset, err = prepareFile(filePath, info.Size, store.mutex)
		if err != nil {
			return nil, err
		}
	}

	upload := &InPlaceUpload{
		FileInfo:     info,
		filePath:     filePath,
		actualOffset: actualOffset,
		parent:       store,
	}

	store.uploads[info.ID] = upload

	return upload, nil
}

func (store *InPlaceDataStore) GetUpload(ctx context.Context, id string) (tusd.Upload, error) {
	if upload, ok := store.uploads[id]; ok {
		return upload, nil
	} else {
		return nil, fmt.Errorf("upload not found")
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

	_, err = file.Seek(upload.actualOffset+offset, io.SeekStart)
	if err != nil {
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
	// Remove the upload from the ID-mapping.
	delete(upload.parent.uploads, upload.filePath)
	return nil
}

func (upload *InPlaceUpload) ConcatUploads(ctx context.Context, uploads []tusd.Upload) (err error) {
	return nil
}

func uid() (string, error) {
	id := make([]byte, uidLength)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(id), err
}

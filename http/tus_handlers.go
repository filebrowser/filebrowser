package fbhttp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/files"
)

// keepUploadActive periodically touches the cache entry to prevent eviction during transfer
func keepUploadActive(cache UploadCache, filePath string) func() {
	stop := make(chan bool)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				cache.Touch(filePath)
			}
		}
	}()

	return func() {
		close(stop)
	}
}

func tusPostHandler(cache UploadCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Create || !d.Check(r.URL.Path) {
			return http.StatusForbidden, nil
		}
		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		switch {
		case errors.Is(err, afero.ErrFileNotFound):
			dirPath := filepath.Dir(r.URL.Path)
			if _, statErr := d.user.Fs.Stat(dirPath); os.IsNotExist(statErr) {
				if mkdirErr := d.user.Fs.MkdirAll(dirPath, d.settings.DirMode); mkdirErr != nil {
					return http.StatusInternalServerError, err
				}
			}
		case err != nil:
			return errToStatus(err), err
		}

		fileFlags := os.O_CREATE | os.O_WRONLY

		// if file exists
		if file != nil {
			if file.IsDir {
				return http.StatusBadRequest, fmt.Errorf("cannot upload to a directory %s", file.RealPath())
			}

			// Existing files will remain untouched unless explicitly instructed to override
			if r.URL.Query().Get("override") != "true" {
				return http.StatusConflict, nil
			}

			// Permission for overwriting the file
			if !d.user.Perm.Modify {
				return http.StatusForbidden, nil
			}

			fileFlags |= os.O_TRUNC
		}

		openFile, err := d.user.Fs.OpenFile(r.URL.Path, fileFlags, d.settings.FileMode)
		if err != nil {
			return errToStatus(err), err
		}
		defer openFile.Close()

		file, err = files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: false,
			Checker:    d,
			Content:    false,
		})
		if err != nil {
			return errToStatus(err), err
		}

		uploadLength, err := getUploadLength(r)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid upload length: %w", err)
		}

		// Enables the user to utilize the PATCH endpoint for uploading file data
		cache.Register(file.RealPath(), uploadLength)

		path, err := url.JoinPath("/", d.server.BaseURL, "/api/tus", r.URL.Path)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid path: %w", err)
		}

		w.Header().Set("Location", path)
		return http.StatusCreated, nil
	})
}

func tusHeadHandler(cache UploadCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		w.Header().Set("Cache-Control", "no-store")
		if !d.user.Perm.Create || !d.Check(r.URL.Path) {
			return http.StatusForbidden, nil
		}

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		if err != nil {
			return errToStatus(err), err
		}

		uploadLength, err := cache.GetLength(file.RealPath())
		if err != nil {
			return http.StatusNotFound, err
		}

		w.Header().Set("Upload-Offset", strconv.FormatInt(file.Size, 10))
		w.Header().Set("Upload-Length", strconv.FormatInt(uploadLength, 10))

		return http.StatusOK, nil
	})
}

func tusPatchHandler(cache UploadCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Create || !d.Check(r.URL.Path) {
			return http.StatusForbidden, nil
		}
		if r.Header.Get("Content-Type") != "application/offset+octet-stream" {
			return http.StatusUnsupportedMediaType, nil
		}

		uploadOffset, err := getUploadOffset(r)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid upload offset")
		}

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})

		switch {
		case errors.Is(err, afero.ErrFileNotFound):
			return http.StatusNotFound, nil
		case err != nil:
			return errToStatus(err), err
		}

		uploadLength, err := cache.GetLength(file.RealPath())
		if err != nil {
			return http.StatusNotFound, err
		}

		// Prevent the upload from being evicted during the transfer
		stop := keepUploadActive(cache, file.RealPath())
		defer stop()

		switch {
		case file.IsDir:
			return http.StatusBadRequest, fmt.Errorf("cannot upload to a directory %s", file.RealPath())
		case file.Size != uploadOffset:
			return http.StatusConflict, fmt.Errorf(
				"%s file size doesn't match the provided offset: %d",
				file.RealPath(),
				uploadOffset,
			)
		}

		openFile, err := d.user.Fs.OpenFile(r.URL.Path, os.O_WRONLY|os.O_APPEND, d.settings.FileMode)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("could not open file: %w", err)
		}
		defer openFile.Close()

		_, err = openFile.Seek(uploadOffset, 0)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("could not seek file: %w", err)
		}

		defer r.Body.Close()
		bytesWritten, err := io.Copy(openFile, r.Body)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("could not write to file: %w", err)
		}

		// Sync the file to ensure all data is written to storage
		// to prevent file corruption.
		if err := openFile.Sync(); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("could not sync file: %w", err)
		}

		newOffset := uploadOffset + bytesWritten
		w.Header().Set("Upload-Offset", strconv.FormatInt(newOffset, 10))

		if newOffset >= uploadLength {
			cache.Complete(file.RealPath())
			_ = d.RunHook(func() error { return nil }, "upload", r.URL.Path, "", d.user)
		}

		return http.StatusNoContent, nil
	})
}

func tusDeleteHandler(cache UploadCache) handleFunc {
	return withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.URL.Path == "/" || !d.user.Perm.Create {
			return http.StatusForbidden, nil
		}

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		if err != nil {
			return errToStatus(err), err
		}

		_, err = cache.GetLength(file.RealPath())
		if err != nil {
			return http.StatusNotFound, err
		}

		err = d.user.Fs.RemoveAll(r.URL.Path)
		if err != nil {
			return errToStatus(err), err
		}

		cache.Complete(file.RealPath())

		return http.StatusNoContent, nil
	})
}

func getUploadLength(r *http.Request) (int64, error) {
	uploadOffset, err := strconv.ParseInt(r.Header.Get("Upload-Length"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid upload length: %w", err)
	}
	return uploadOffset, nil
}

func getUploadOffset(r *http.Request) (int64, error) {
	uploadOffset, err := strconv.ParseInt(r.Header.Get("Upload-Offset"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid upload offset: %w", err)
	}
	return uploadOffset, nil
}

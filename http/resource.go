package http

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/fileutils"
)

var resourceGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file, err := files.NewFileInfo(files.FileOptions{
		Fs:         d.user.Fs,
		Path:       r.URL.Path,
		Modify:     d.user.Perm.Modify,
		Expand:     true,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
	})
	if err != nil {
		return errToStatus(err), err
	}

	if file.IsDir {
		file.Listing.Sorting = d.user.Sorting
		file.Listing.ApplySort()
		return renderJSON(w, r, file)
	}

	if checksum := r.URL.Query().Get("checksum"); checksum != "" {
		err := file.Checksum(checksum)
		if err == errors.ErrInvalidOption {
			return http.StatusBadRequest, nil
		} else if err != nil {
			return http.StatusInternalServerError, err
		}

		// do not waste bandwidth if we just want the checksum
		file.Content = ""
	}

	return renderJSON(w, r, file)
})

func resourceDeleteHandler(fileCache FileCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.URL.Path == "/" || !d.user.Perm.Delete {
			return http.StatusForbidden, nil
		}

		file, err := files.NewFileInfo(files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     true,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		if err != nil {
			return errToStatus(err), err
		}

		// delete thumbnails
		for _, previewSizeName := range PreviewSizeNames() {
			size, _ := ParsePreviewSize(previewSizeName)
			if err := fileCache.Delete(r.Context(), previewCacheKey(file.Path, size)); err != nil { //nolint:govet
				return errToStatus(err), err
			}
		}

		err = d.RunHook(func() error {
			return d.user.Fs.RemoveAll(r.URL.Path)
		}, "delete", r.URL.Path, "", d.user)

		if err != nil {
			return errToStatus(err), err
		}

		return http.StatusOK, nil
	})
}

var resourcePostHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Create {
		return http.StatusForbidden, nil
	}

	defer func() {
		_, _ = io.Copy(ioutil.Discard, r.Body)
	}()

	// Directories creation on POST.
	if strings.HasSuffix(r.URL.Path, "/") {
		err := d.user.Fs.MkdirAll(r.URL.Path, 0775)
		return errToStatus(err), err
	}

	if r.URL.Query().Get("override") != "true" {
		if _, err := d.user.Fs.Stat(r.URL.Path); err == nil {
			return http.StatusConflict, nil
		}
	}

	err := d.RunHook(func() error {
		info, _ := writeFile(d.user.Fs, r.URL.Path, r.Body)

		etag := fmt.Sprintf(`"%x%x"`, info.ModTime().UnixNano(), info.Size())
		w.Header().Set("ETag", etag)
		return nil
	}, "upload", r.URL.Path, "", d.user)

	if err != nil {
		_ = d.user.Fs.RemoveAll(r.URL.Path)
	}

	return errToStatus(err), err
})

var resourcePutHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Modify {
		return http.StatusForbidden, nil
	}

	defer func() {
		_, _ = io.Copy(ioutil.Discard, r.Body)
	}()

	// Only allow PUT for files.
	if strings.HasSuffix(r.URL.Path, "/") {
		return http.StatusMethodNotAllowed, nil
	}

	err := d.RunHook(func() error {
		info, _ := writeFile(d.user.Fs, r.URL.Path, r.Body)

		etag := fmt.Sprintf(`"%x%x"`, info.ModTime().UnixNano(), info.Size())
		w.Header().Set("ETag", etag)
		return nil
	}, "save", r.URL.Path, "", d.user)

	if err != nil {
		_ = d.user.Fs.RemoveAll(r.URL.Path)
	}

	return errToStatus(err), err
})

var resourcePatchHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	src := r.URL.Path
	dst := r.URL.Query().Get("destination")
	action := r.URL.Query().Get("action")
	dst, err := url.QueryUnescape(dst)
	if err != nil {
		return errToStatus(err), err
	}
	if dst == "/" || src == "/" {
		return http.StatusForbidden, nil
	}
	if err = checkParent(src, dst); err != nil {
		return http.StatusBadRequest, err
	}

	override := r.URL.Query().Get("override") == "true"
	rename := r.URL.Query().Get("rename") == "true"
	if !override && !rename {
		if _, err = d.user.Fs.Stat(dst); err == nil {
			return http.StatusConflict, nil
		}
	}
	if rename {
		dst = addVersionSuffix(dst, d.user.Fs)
	}

	err = d.RunHook(func() error {
		switch action {
		// TODO: use enum
		case "copy":
			if !d.user.Perm.Create {
				return errors.ErrPermissionDenied
			}

			return fileutils.Copy(d.user.Fs, src, dst)
		case "rename":
			if !d.user.Perm.Rename {
				return errors.ErrPermissionDenied
			}
			src = path.Clean("/" + src)
			dst = path.Clean("/" + dst)

			return fileutils.MoveFile(d.user.Fs, src, dst)
		default:
			return fmt.Errorf("unsupported action %s: %w", action, errors.ErrInvalidRequestParams)
		}
	}, action, src, dst, d.user)

	return errToStatus(err), err
})

func checkParent(src, dst string) error {
	rel, err := filepath.Rel(src, dst)
	if err != nil {
		return err
	}

	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, "../") && rel != ".." && rel != "." {
		return errors.ErrSourceIsParent
	}

	return nil
}

func addVersionSuffix(source string, fs afero.Fs) string {
	counter := 1
	dir, name := path.Split(source)
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	for {
		if _, err := fs.Stat(source); err != nil {
			break
		}
		renamed := fmt.Sprintf("%s(%d)%s", base, counter, ext)
		source = path.Join(dir, renamed)
		counter++
	}

	return source
}

func writeFile(fs afero.Fs, dst string, in io.Reader) (os.FileInfo, error) {
	dir, _ := path.Split(dst)
	err := fs.MkdirAll(dir, 0775)
	if err != nil {
		return nil, err
	}

	file, err := fs.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(file, in)
	if err != nil {
		return nil, err
	}

	// Gets the info about the file.
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return info, nil
}

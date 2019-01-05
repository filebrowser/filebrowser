package http

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/filebrowser/filebrowser/v2/files"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/fileutils"
)

var resourceGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file, err := files.NewFileInfo(files.FileOptions{
		Fs:      d.user.Fs,
		Path:    r.URL.Path,
		Modify:  d.user.Perm.Modify,
		Expand:  true,
		Checker: d,
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

var resourceDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if r.URL.Path == "/" || !d.user.Perm.Delete {
		return http.StatusForbidden, nil
	}

	err := d.RunHook(func() error {
		return d.user.Fs.RemoveAll(r.URL.Path)
	}, "delete", r.URL.Path, "", d.user)

	if err != nil {
		return errToStatus(err), err
	}

	return http.StatusOK, nil
})

var resourcePostPutHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Create && r.Method == http.MethodPost {
		return http.StatusForbidden, nil
	}

	if !d.user.Perm.Modify && r.Method == http.MethodPut {
		return http.StatusForbidden, nil
	}

	defer func() {
		io.Copy(ioutil.Discard, r.Body)
	}()

	// For directories, only allow POST for creation.
	if strings.HasSuffix(r.URL.Path, "/") {
		if r.Method == http.MethodPut {
			return http.StatusMethodNotAllowed, nil
		}

		err := d.user.Fs.MkdirAll(r.URL.Path, 0775)
		return errToStatus(err), err
	}

	if r.Method == http.MethodPost && r.URL.Query().Get("override") != "true" {
		if _, err := d.user.Fs.Stat(r.URL.Path); err == nil {
			return http.StatusConflict, nil
		}
	}

	err := d.RunHook(func() error {
		file, err := d.user.Fs.OpenFile(r.URL.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, r.Body)
		if err != nil {
			return err
		}

		// Gets the info about the file.
		info, err := file.Stat()
		if err != nil {
			return err
		}

		etag := fmt.Sprintf(`"%x%x"`, info.ModTime().UnixNano(), info.Size())
		w.Header().Set("ETag", etag)
		return nil
	}, "upload", r.URL.Path, "", d.user)

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

	switch action {
	case "copy":
		if !d.user.Perm.Create {
			return http.StatusForbidden, nil
		}
	case "rename":
	default:
		action = "rename"
		if !d.user.Perm.Rename {
			return http.StatusForbidden, nil
		}
	}

	err = d.RunHook(func() error {
		if action == "copy" {
			return fileutils.Copy(d.user.Fs, src, dst)
		}

		return d.user.Fs.Rename(src, dst)
	}, action, src, dst, d.user)

	return errToStatus(err), err
})

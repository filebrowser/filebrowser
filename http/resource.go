package http

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/filebrowser/filebrowser/fileutils"
	"github.com/filebrowser/filebrowser/lib"
)

const apiResourcePrefix = "/api/resources"

func httpFsErr(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case os.IsPermission(err):
		return http.StatusForbidden
	case os.IsNotExist(err):
		return http.StatusNotFound
	case os.IsExist(err):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func (e *Env) getResourceData(w http.ResponseWriter, r *http.Request, prefix string) (string, *lib.User, bool) {
	user, ok := e.getUser(w, r)
	if !ok {
		return "", nil, ok
	}

	path := strings.TrimPrefix(r.URL.Path, prefix)
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}

	return path, user, true
}

func (e *Env) resourceGetHandler(w http.ResponseWriter, r *http.Request) {
	path, user, ok := e.getResourceData(w, r, apiResourcePrefix)
	if !ok {
		return
	}

	file, err := e.NewFile(path, user)
	if err != nil {
		httpErr(w, r, httpFsErr(err), err)
		return
	}

	if file.IsDir {
		file.Listing.Sorting = user.Sorting
		file.Listing.ApplySort()
		renderJSON(w, r, file)
		return
	}

	if !user.Perm.Modify && file.Type == "text" {
		// TODO: move to detet file type
		file.Type = "textImmutable"
	}

	if checksum := r.URL.Query().Get("checksum"); checksum != "" {
		err = e.Checksum(file,user, checksum)
		if err == lib.ErrInvalidOption {
			httpErr(w, r, http.StatusBadRequest, nil)
			return
		} else if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
			return
		}

		// do not waste bandwidth if we just want the checksum
		file.Content = ""
	}

	renderJSON(w, r, file)
}

func (e *Env) resourceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	path, user, ok := e.getResourceData(w, r, apiResourcePrefix)
	if !ok {
		return
	}

	if path == "/" || !user.Perm.Delete {
		httpErr(w, r, http.StatusForbidden, nil)
		return
	}

	err := e.RunHook(func() error {
		return user.Fs.RemoveAll(path)
	}, "delete", path, "", user)

	if err != nil {
		httpErr(w, r, httpFsErr(err), err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (e *Env) resourcePostPutHandler(w http.ResponseWriter, r *http.Request) {
	path, user, ok := e.getResourceData(w, r, apiResourcePrefix)
	if !ok {
		return
	}

	if !user.Perm.Create && r.Method == http.MethodPost {
		httpErr(w, r, http.StatusForbidden, nil)
		return
	}

	if !user.Perm.Modify && r.Method == http.MethodPut {
		httpErr(w, r, http.StatusForbidden, nil)
		return
	}

	defer func() {
		io.Copy(ioutil.Discard, r.Body)
	}()

	// For directories, only allow POST for creation.
	if strings.HasSuffix(r.URL.Path, "/") {
		if r.Method == http.MethodPut {
			httpErr(w, r, http.StatusMethodNotAllowed, nil)
		} else {
			err := user.Fs.MkdirAll(path, 0775)
			httpErr(w, r, httpFsErr(err), err)
		}

		return
	}

	if r.Method == http.MethodPost && r.URL.Query().Get("override") != "true" {
		if _, err := user.Fs.Stat(path); err == nil {
			httpErr(w, r, http.StatusConflict, nil)
			return
		}
	}

	err := e.RunHook(func() error {
		file, err := user.Fs.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
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
	}, "upload", path, "", user)

	if err != nil {
		httpErr(w, r, httpFsErr(err), err)
		return
	}

	httpErr(w, r, http.StatusOK, nil)
}

func (e *Env) resourcePatchHandler(w http.ResponseWriter, r *http.Request) {
	src, user, ok := e.getResourceData(w, r, apiResourcePrefix)
	if !ok {
		return
	}

	dst := r.URL.Query().Get("destination")
	action := r.URL.Query().Get("action")
	dst, err := url.QueryUnescape(dst)

	if err != nil {
		httpErr(w, r, httpFsErr(err), err)
		return
	}

	if dst == "/" || src == "/" {
		httpErr(w, r, http.StatusForbidden, nil)
		return
	}

	switch action {
	case "copy":
		if !user.Perm.Create {
			httpErr(w, r, http.StatusForbidden, nil)
			return
		}
	case "rename":
	default:
		action = "rename"
		if !user.Perm.Rename {
			httpErr(w, r, http.StatusForbidden, nil)
			return
		}
	}

	err = e.RunHook(func() error {
		if action == "copy" {
			return fileutils.Copy(user.Fs, src, dst)
		}

		return user.Fs.Rename(src, dst)
	}, action, src, dst, user)

	httpErr(w, r, httpFsErr(err), err)
}

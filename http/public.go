package http

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/filebrowser/filebrowser/v2/files"
	libErrors "github.com/filebrowser/filebrowser/v2/errors"
)

var withHashFile = func(fn handleFunc) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		id, rootName, path := ifPathWithName(r)
		link, err := d.store.Share.GetByHash(id)
		if err != nil {
			return errToStatus(err), err
		}

		if rootName != "" && filepath.Base(rootName) != filepath.Base(link.Path) {
			return errToStatus(libErrors.ErrNotExist), libErrors.ErrNotExist
		}

		user, err := d.store.Users.Get(d.server.Root, link.UserID)
		if err != nil {
			return errToStatus(err), err
		}

		d.user = user

		file, err := files.NewFileInfo(files.FileOptions{
			Fs:      d.user.Fs,
			Path:    filepath.Join(link.Path, path),
			Modify:  d.user.Perm.Modify,
			Expand:  true,
			Checker: d,
		})
		if err != nil {
			return errToStatus(err), err
		}

		d.raw = file
		return fn(w, r, d)
	}
}

// ref to https://github.com/filebrowser/filebrowser/pull/727
// `/api/public/dl/MEEuZK-v/file-name.txt` for old browsers to save file with correct name
func ifPathWithName(r *http.Request) (id, rootName, path string) {
	pathElements := strings.Split(r.URL.Path, "/")
	// prevent maliciously constructed parameters like `/api/public/dl/XZzCDnK2_not_exists_hash_name`
	// len(pathElements) will be 1, and golang will panic `runtime error: index out of range`
	switch len(pathElements) {
	case 1:
		return r.URL.Path, "", ""
	case 2:
		return pathElements[0], pathElements[1], ""
	default:
		return pathElements[0], pathElements[1], strings.Join(pathElements[2:], "/")
	}
}

var publicShareHandler = withHashFile(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file := d.raw.(*files.FileInfo)

	if file.IsDir {
		file.Listing.Sorting = files.Sorting{By: "name", Asc: false}
		file.Listing.ApplySort()
		return renderJSON(w, r, file)
	}

	return renderJSON(w, r, file)
})

var publicDlHandler = withHashFile(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file := d.raw.(*files.FileInfo)
	if !file.IsDir {
		return rawFileHandler(w, r, file)
	}

	return rawDirHandler(w, r, d, file)
})

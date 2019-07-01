package http

import (
	"net/http"
	"strings"

	"github.com/filebrowser/filebrowser/v2/files"
)

var withHashFile = func(fn handleFunc) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		link, err := d.store.Share.GetByHash(r.URL.Path)
		if err != nil {
			link, err = d.store.Share.GetByHash(ifPathWithName(r))
			if err != nil {
				return errToStatus(err), err
			}
		}

		user, err := d.store.Users.Get(d.server.Root, link.UserID)
		if err != nil {
			return errToStatus(err), err
		}

		d.user = user

		file, err := files.NewFileInfo(files.FileOptions{
			Fs:      d.user.Fs,
			Path:    link.Path,
			Modify:  d.user.Perm.Modify,
			Expand:  false,
			Checker: d,
		})
		if err != nil {
			return errToStatus(err), err
		}

		d.raw = file
		return fn(w, r, d)
	}
}

func ifPathWithName(r *http.Request) string {
	pathElements := strings.Split(r.URL.Path, "/")
	id := pathElements[len(pathElements)-2]
	return id
}

var publicShareHandler = withHashFile(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	return renderJSON(w, r, d.raw)
})

var publicDlHandler = withHashFile(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file := d.raw.(*files.FileInfo)
	if !file.IsDir {
		return rawFileHandler(w, r, file)
	}

	return rawDirHandler(w, r, d, file)
})

var publicShareFolderHandler = withHashFile(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	baseFolder := d.raw.(*files.FileInfo)

	file, err := files.NewFileInfo(files.FileOptions{
		Fs:      d.user.Fs,
		Path:    baseFolder.Path + "/" + r.Header.Get("Relative-Path"),
		Modify:  d.user.Perm.Modify,
		Expand:  true,
		Checker: d,
	})
	if err != nil {
		return errToStatus(err), err
	}

	file.Listing.Sorting = d.user.Sorting
	file.Listing.ApplySort()
	return renderJSON(w, r, file)
})

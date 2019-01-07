package http

import (
	"net/http"

	"github.com/filebrowser/filebrowser/v2/files"
)

var withHashFile = func(fn handleFunc) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		link, err := d.store.Share.GetByHash(r.URL.Path)
		if err != nil {
			return errToStatus(err), err
		}

		user, err := d.store.Users.Get(d.settings.Root, link.UserID)
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

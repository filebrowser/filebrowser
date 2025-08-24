package http

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/filebrowser/filebrowser/v2/files"
)

type downloadBody struct {
	URL  string `json:"url"`
	Path string `json:"path"`
}

var urlDownloadHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Create {
		return http.StatusForbidden, nil
	}

	var body downloadBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if body.URL == "" || body.Path == "" {
		return http.StatusBadRequest, nil
	}

	_, err = url.ParseRequestURI(body.URL)
	if err != nil {
		return http.StatusBadRequest, err
	}

	fileName := path.Base(body.URL)
	filePath := path.Join(body.Path, fileName)

	if !d.Check(filePath) {
		return http.StatusForbidden, nil
	}

	resp, err := http.Get(body.URL)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	err = d.RunHook(func() error {
		_, writeErr := writeFile(d.user.Fs, filePath, resp.Body, d.settings.FileMode, d.settings.DirMode)
		return writeErr
	}, "upload", filePath, "", d.user)

	if err != nil {
		_ = d.user.Fs.RemoveAll(filePath)
		return http.StatusInternalServerError, err
	}

	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.user.Fs,
		Path:       filePath,
		Modify:     d.user.Perm.Modify,
		Expand:     false,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, file)
})

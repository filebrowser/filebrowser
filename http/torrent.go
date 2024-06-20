package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/users"
)

func withPermTorrent(fn handleFunc) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Torrent {
			return http.StatusForbidden, nil
		}

		return fn(w, r, d)
	})
}

var torrentGetHandler = withPermTorrent(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// return default torrent options
	s, err := d.GetDefaultCreateBody(d.user.CreateBody)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to get default create body: %w", err)
	}

	return renderJSON(w, r, s)
})

var torrentPostHandler = withPermTorrent(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.user.Fs,
		Path:       r.URL.Path,
		Modify:     d.user.Perm.Modify,
		Expand:     true,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
		Content:    true,
	})
	if err != nil {
		return errToStatus(err), err
	}
	fPath := file.RealPath()

	var body users.CreateTorrentBody
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return http.StatusBadRequest, fmt.Errorf("failed to decode body: %w", err)
		}
		defer r.Body.Close()
	}

	err = d.Torrent.MakeTorrent(fPath, body)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	d.user.CreateBody = body

	err = d.store.Users.Update(d.user)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, nil)
})

var publishPostHandler = withPermTorrent(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.user.Fs,
		Path:       r.URL.Path,
		Modify:     d.user.Perm.Modify,
		Expand:     true,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
		Content:    true,
	})
	if err != nil {
		return errToStatus(err), err
	}
	tPath := file.RealPath()
	// only folder path
	fPath := filepath.Dir(tPath)

	torrentPath := tPath
	savePath := fPath

	err = d.Torrent.PublishTorrent(torrentPath, savePath)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, nil)
})

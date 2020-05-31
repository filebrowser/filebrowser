package http

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/share"
)

func withPermShare(fn handleFunc) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Share {
			return http.StatusForbidden, nil
		}

		return fn(w, r, d)
	})
}

var shareGetsHandler = withPermShare(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	s, err := d.store.Share.Gets(r.URL.Path, d.user.ID)
	if err == errors.ErrNotExist {
		return renderJSON(w, r, []*share.Link{})
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, s)
})

var shareDeleteHandler = withPermShare(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	hash := strings.TrimSuffix(r.URL.Path, "/")
	hash = strings.TrimPrefix(hash, "/")

	if hash == "" {
		return http.StatusBadRequest, nil
	}

	err := d.store.Share.Delete(hash)
	return errToStatus(err), err
})

var sharePostHandler = withPermShare(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	var s *share.Link
	rawExpire := r.URL.Query().Get("expires")
	unit := r.URL.Query().Get("unit")

	if rawExpire == "" {
		var err error
		s, err = d.store.Share.GetPermanent(r.URL.Path, d.user.ID)
		if err == nil {
			if _, err := w.Write([]byte(path.Join(d.server.BaseURL, "/share/", s.Hash))); err != nil {
				return http.StatusInternalServerError, err
			}
			return 0, nil
		}
	}

	bytes := make([]byte, 6)
	_, err := rand.Read(bytes)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	str := base64.URLEncoding.EncodeToString(bytes)

	var expire int64 = 0

	if rawExpire != "" {
		num, err := strconv.Atoi(rawExpire)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		var add time.Duration
		switch unit {
		case "seconds":
			add = time.Second * time.Duration(num)
		case "minutes":
			add = time.Minute * time.Duration(num)
		case "days":
			add = time.Hour * 24 * time.Duration(num)
		default:
			add = time.Hour * time.Duration(num)
		}

		expire = time.Now().Add(add).Unix()
	}

	s = &share.Link{
		Path:   r.URL.Path,
		Hash:   str,
		Expire: expire,
		UserID: d.user.ID,
	}

	if err := d.store.Share.Save(s); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, s)
})

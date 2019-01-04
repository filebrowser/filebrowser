package http

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/errors"
	"github.com/filebrowser/filebrowser/share"
)

func withPermShare(fn handleFunc) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Share {
			return http.StatusForbidden, nil
		}

		return fn(w, r, d)
	})
}

var shareGetHandler = withPermShare(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	s, err := d.store.Share.Gets(r.URL.Path)
	if err == errors.ErrNotExist {
		return renderJSON(w, r, []*share.Link{})
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	for i, link := range s {
		if link.Expires && link.ExpireDate.Before(time.Now()) {
			d.store.Share.Delete(link.Hash)
			s = append(s[:i], s[i+1:]...)
		}
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
	expire := r.URL.Query().Get("expires")
	unit := r.URL.Query().Get("unit")

	if expire == "" {
		var err error
		s, err = d.store.Share.GetPermanent(r.URL.Path)
		if err == nil {
			w.Write([]byte(d.settings.BaseURL + "/share/" + s.Hash))
			return 0, nil
		}
	}

	bytes := make([]byte, 6)
	_, err := rand.Read(bytes)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	str := base64.URLEncoding.EncodeToString(bytes)

	s = &share.Link{
		Path:    r.URL.Path,
		Hash:    str,
		Expires: expire != "",
	}

	if expire != "" {
		num, err := strconv.Atoi(expire)
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

		s.ExpireDate = time.Now().Add(add)
	}

	if err := d.store.Share.Save(s); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, s)
})

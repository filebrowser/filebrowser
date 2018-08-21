package http

import (
	"encoding/base64"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	fb "github.com/filebrowser/filebrowser/lib"
)

func shareHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	r.URL.Path = sanitizeURL(r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		return shareGetHandler(c, w, r)
	case http.MethodDelete:
		return shareDeleteHandler(c, w, r)
	case http.MethodPost:
		return sharePostHandler(c, w, r)
	}

	return http.StatusNotImplemented, nil
}

func shareGetHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	path := filepath.Join(c.User.Scope, r.URL.Path)
	s, err := c.Store.Share.GetByPath(path)
	if err == fb.ErrNotExist {
		return http.StatusNotFound, nil
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	for i, link := range s {
		if link.Expires && link.ExpireDate.Before(time.Now()) {
			c.Store.Share.Delete(link.Hash)
			s = append(s[:i], s[i+1:]...)
		}
	}

	if len(s) == 0 {
		return http.StatusNotFound, nil
	}

	return renderJSON(w, s)
}

func sharePostHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	path := filepath.Join(c.User.Scope, r.URL.Path)

	var s *fb.ShareLink
	expire := r.URL.Query().Get("expires")
	unit := r.URL.Query().Get("unit")

	if expire == "" {
		var err error
		s, err = c.Store.Share.GetPermanent(path)
		if err == nil {
			w.Write([]byte(c.RootURL() + "/share/" + s.Hash))
			return 0, nil
		}
	}

	bytes, err := fb.GenerateRandomBytes(6)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	str := base64.URLEncoding.EncodeToString(bytes)

	s = &fb.ShareLink{
		Path:    path,
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

	if err := c.Store.Share.Save(s); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, s)
}

func shareDeleteHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	s, err := c.Store.Share.Get(strings.TrimPrefix(r.URL.Path, "/"))
	if err == fb.ErrNotExist {
		return http.StatusNotFound, nil
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = c.Store.Share.Delete(s.Hash)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

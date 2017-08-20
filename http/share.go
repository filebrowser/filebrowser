package http

import (
	"encoding/hex"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/asdine/storm"
	fm "github.com/hacdias/filemanager"
)

func shareHandler(c *fm.Context, w http.ResponseWriter, r *http.Request) (int, error) {
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

func shareGetHandler(c *fm.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	path := filepath.Join(string(c.User.FileSystem), r.URL.Path)
	s, err := c.Store.Share.GetByPath(path)
	if err == storm.ErrNotFound {
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

	return renderJSON(w, s)
}

func sharePostHandler(c *fm.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	path := filepath.Join(string(c.User.FileSystem), r.URL.Path)

	var s *fm.ShareLink
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

	bytes, err := fm.GenerateRandomBytes(32)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	str := hex.EncodeToString(bytes)

	s = &fm.ShareLink{
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

func shareDeleteHandler(c *fm.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	s, err := c.Store.Share.Get(strings.TrimPrefix(r.URL.Path, "/"))
	if err == storm.ErrNotFound {
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

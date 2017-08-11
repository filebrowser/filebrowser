package filemanager

import (
	"encoding/hex"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

type shareLink struct {
	Hash       string    `json:"hash" storm:"id,index"`
	Path       string    `json:"path" storm:"index"`
	Expires    bool      `json:"expires"`
	ExpireDate time.Time `json:"expireDate"`
}

func shareHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
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

func shareGetHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var (
		s    []*shareLink
		path = filepath.Join(string(c.User.FileSystem), r.URL.Path)
	)

	err := c.db.Find("Path", path, &s)
	if err == storm.ErrNotFound {
		return http.StatusNotFound, nil
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	for i, link := range s {
		if link.Expires && link.ExpireDate.Before(time.Now()) {
			c.db.DeleteStruct(&shareLink{Hash: link.Hash})
			s = append(s[:i], s[i+1:]...)
		}
	}

	return renderJSON(w, s)
}

func sharePostHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	path := filepath.Join(string(c.User.FileSystem), r.URL.Path)

	var s shareLink
	expire := r.URL.Query().Get("expires")
	unit := r.URL.Query().Get("unit")

	if expire == "" {
		err := c.db.Select(q.Eq("Path", path), q.Eq("Expires", false)).First(&s)
		if err == nil {
			w.Write([]byte(c.RootURL() + "/share/" + s.Hash))
			return 0, nil
		}
	}

	bytes, err := generateRandomBytes(32)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	str := hex.EncodeToString(bytes)

	s = shareLink{
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

	err = c.db.Save(&s)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, s)
}

func shareDeleteHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var s shareLink

	err := c.db.One("Hash", strings.TrimPrefix(r.URL.Path, "/"), &s)
	if err == storm.ErrNotFound {
		return http.StatusNotFound, nil
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = c.db.DeleteStruct(&s)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

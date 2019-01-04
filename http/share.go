package http

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/lib"
)

const apiSharePrefix = "/api/share"

func (e *Env) getShareData(w http.ResponseWriter, r *http.Request, prefix string) (string, bool) {
	relPath, user, ok := e.getResourceData(w, r, apiSharePrefix)
	if !ok {
		return "", false
	}

	if !user.Perm.Share {
		httpErr(w, r, http.StatusForbidden, nil)
		return "", false
	}

	return user.FullPath(relPath), ok
}

func (e *Env) shareGetHandler(w http.ResponseWriter, r *http.Request) {
	path, ok := e.getShareData(w, r, apiSharePrefix)
	if !ok {
		return
	}

	s, err := e.GetLinksByPath(path)
	if err == lib.ErrNotExist {
		renderJSON(w, r, []*lib.ShareLink{})
		return
	}

	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	for i, link := range s {
		if link.Expires && link.ExpireDate.Before(time.Now()) {
			e.DeleteLink(link.Hash)
			s = append(s[:i], s[i+1:]...)
		}
	}

	renderJSON(w, r, s)
}

func (e *Env) shareDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := e.getUser(w, r)
	if !ok {
		return
	}

	if !user.Perm.Share {
		httpErr(w, r, http.StatusForbidden, nil)
		return
	}

	hash := strings.TrimPrefix(r.URL.Path, apiSharePrefix)
	hash = strings.TrimSuffix(hash, "/")
	hash = strings.TrimPrefix(hash, "/")
	if hash == "" {
		return
	}

	err := e.DeleteLink(hash)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (e *Env) sharePostHandler(w http.ResponseWriter, r *http.Request) {
	path, ok := e.getShareData(w, r, apiSharePrefix)
	if !ok {
		return
	}

	var s *lib.ShareLink
	expire := r.URL.Query().Get("expires")
	unit := r.URL.Query().Get("unit")

	if expire == "" {
		var err error
		s, err = e.GetLinkPermanent(path)
		if err == nil {
			w.Write([]byte(e.GetSettings().BaseURL + "/share/" + s.Hash))
			return
		}
	}

	bytes := make([]byte, 6)
	_, err := rand.Read(bytes)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	str := base64.URLEncoding.EncodeToString(bytes)

	s = &lib.ShareLink{
		Path:    path,
		Hash:    str,
		Expires: expire != "",
	}

	if expire != "" {
		num, err := strconv.Atoi(expire)
		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
			return
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

	if err := e.SaveLink(s); err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	renderJSON(w, r, s)
}

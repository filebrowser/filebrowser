package http

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/filebrowser/filebrowser/types"
	"github.com/gorilla/mux"
)

func getUserID(r *http.Request) (uint, error) {
	vars := mux.Vars(r)
	i, err := strconv.ParseUint(vars["id"], 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(i), err
}

func (e *Env) usersGetHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := e.getUser(w, r)
	if !ok {
		return
	}

	if !user.Perm.Admin {
		httpErr(w, http.StatusForbidden, nil)
		return
	}

	users, err := e.Store.Users.Gets()
	if err != nil {
		httpErr(w, http.StatusInternalServerError, err)
		return
	}

	for _, u := range users {
		u.Password = ""
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	renderJSON(w, users)
}

func (e *Env) userSelfOrAdmin(w http.ResponseWriter, r *http.Request) (*types.User, uint, bool) {
	user, ok := e.getUser(w, r)
	if !ok {
		return nil, 0, false
	}

	id, err := getUserID(r)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, err)
		return nil, 0, false
	}

	if user.ID != id && !user.Perm.Admin {
		httpErr(w, http.StatusForbidden, nil)
		return nil, 0, false
	}

	return user, id, true
}

func (e *Env) userGetHandler(w http.ResponseWriter, r *http.Request) {
	_, id, ok := e.userSelfOrAdmin(w, r)
	if !ok {
		return
	}

	u, err := e.Store.Users.Get(id)
	if err == types.ErrNotExist {
		httpErr(w, http.StatusNotFound, nil)
		return
	}

	if err != nil {
		httpErr(w, http.StatusInternalServerError, err)
		return
	}

	u.Password = ""
	renderJSON(w, u)
}

func (e *Env) userDeleteHandler(w http.ResponseWriter, r *http.Request) {
	_, id, ok := e.userSelfOrAdmin(w, r)
	if !ok {
		return
	}

	err := e.Store.Users.Delete(id)
	if err == types.ErrNotExist {
		httpErr(w, http.StatusNotFound, nil)
		return
	}

	if err != nil {
		httpErr(w, http.StatusInternalServerError, err)
	}
}

func (e *Env) userPostHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: fill me
}

func (e *Env) userPutHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: fill me
}

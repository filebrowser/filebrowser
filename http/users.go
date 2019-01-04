package http

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/filebrowser/filebrowser/errors"
	"github.com/filebrowser/filebrowser/users"
	"github.com/gorilla/mux"
)

type modifyUserRequest struct {
	modifyRequest
	Data *users.User `json:"data"`
}

func getUserID(r *http.Request) (uint, error) {
	vars := mux.Vars(r)
	i, err := strconv.ParseUint(vars["id"], 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(i), err
}

func getUser(w http.ResponseWriter, r *http.Request) (*modifyUserRequest, error) {
	if r.Body == nil {
		return nil, errors.ErrEmptyRequest
	}

	req := &modifyUserRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return nil, err
	}

	if req.What != "user" {
		return nil, errors.ErrInvalidDataType
	}

	return req, nil
}

func withSelfOrAdmin(fn handleFunc) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		id, err := getUserID(r)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if d.user.ID != id && !d.user.Perm.Admin {
			return http.StatusForbidden, nil
		}

		d.raw = id
		return fn(w, r, d)
	})
}

var usersGetHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	users, err := d.store.Users.Gets()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	
	for _, u := range users {
		u.Password = ""
	}
	
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})
	
	return renderJSON(w, r, users)
})

var userGetHandler = withSelfOrAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	u, err := d.store.Users.Get(d.raw.(uint))
	if err == errors.ErrNotExist {
		return http.StatusNotFound, err
	}
	
	if err != nil {
		return http.StatusInternalServerError, err
	}

	u.Password = ""
	return renderJSON(w, r, u)
})

var userDeleteHandler = withSelfOrAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	err := d.store.Users.Delete(d.raw.(uint))
	if err == errors.ErrNotExist {
		return http.StatusNotFound, err
	}
	
	return http.StatusOK, nil
})

/*

func (e *env) userPostHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := e.getAdminUser(w, r)
	if !ok {
		return
	}

	req, ok := getUser(w, r)
	if !ok {
		return
	}

	if len(req.Which) != 0 {
		httpErr(w, r, http.StatusBadRequest, nil)
		return
	}

	if req.Data.Password == "" {
		httpErr(w, r, http.StatusBadRequest, lib.ErrEmptyPassword)
		return
	}

	var err error
	req.Data.Password, err = lib.HashPwd(req.Data.Password)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	err = e.SaveUser(req.Data)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Location", "/settings/users/"+strconv.FormatUint(uint64(req.Data.ID), 10))
	w.WriteHeader(http.StatusCreated)
}

func (e *env) userPutHandler(w http.ResponseWriter, r *http.Request) {
	sessionUser, modifiedID, ok := e.userSelfOrAdmin(w, r)
	if !ok {
		return
	}

	req, ok := getUser(w, r)
	if !ok {
		return
	}

	if req.Data.ID != modifiedID {
		httpErr(w, r, http.StatusBadRequest, nil)
		return
	}

	var err error

	if len(req.Which) == 1 && req.Which[0] == "all" {
		if !sessionUser.Perm.Admin {
			httpErr(w, r, http.StatusForbidden, nil)
			return
		}

		if req.Data.Password != "" {
			req.Data.Password, err = lib.HashPwd(req.Data.Password)
		} else {
			var suser *users.User
			suser, err = e.GetUser(modifiedID)
			req.Data.Password = suser.Password
		}

		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
			return
		}

		req.Which = []string{}
	}

	for k, v := range req.Which {
		if v == "password" {
			if !sessionUser.Perm.Admin && sessionUser.LockPassword {
				httpErr(w, r, http.StatusForbidden, nil)
				return
			}

			req.Data.Password, err = lib.HashPwd(req.Data.Password)
			if err != nil {
				httpErr(w, r, http.StatusInternalServerError, err)
				return
			}
		}

		if !sessionUser.Perm.Admin && (v == "scope" || v == "perm" || v == "username") {
			httpErr(w, r, http.StatusForbidden, nil)
			return
		}

		req.Which[k] = strings.Title(v)
	}

	err = e.UpdateUser(req.Data, req.Which...)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
	}
} */

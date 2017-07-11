package filemanager

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/asdine/storm"
)

func usersHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	switch r.Method {
	case http.MethodGet:
		return usersGetHandler(c, w, r)
	case http.MethodPost:
		return usersPostHandler(c, w, r)
	case http.MethodDelete:
		return usersDeleteHandler(c, w, r)
	case http.MethodPut:
		return usersPutHandler(c, w, r)
	}

	return http.StatusNotImplemented, nil
}

// usersGetHandler is used to handle the GET requests for /api/users. It can print a list
// of users or a specific user. The password hash is always removed before being sent to the
// client.
func usersGetHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}

	// If the request is a list of users.
	if r.URL.Path == "/" {
		users := []User{}

		for _, user := range c.FM.Users {
			// Copies the user and removes the password.
			u := *user
			u.Password = ""
			users = append(users, u)
		}

		sort.Slice(users, func(i, j int) bool {
			return users[i].ID < users[j].ID
		})

		return renderJSON(w, users)
	}

	// Otherwise we just want one, specific, user.
	sid := strings.TrimPrefix(r.URL.Path, "/")
	sid = strings.TrimSuffix(sid, "/")

	id, err := strconv.Atoi(sid)
	if err != nil {
		return http.StatusNotFound, err
	}

	// Searches for the user and prints the one who matches.
	for _, user := range c.FM.Users {
		if user.ID != id {
			continue
		}

		u := *user
		u.Password = ""
		return renderJSON(w, u)
	}

	// If there aren't any matches, return Not Found.
	return http.StatusNotFound, nil
}

func usersPostHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}

	// New users should be created on /api/users.
	if r.URL.Path != "/" {
		return http.StatusMethodNotAllowed, nil
	}

	// If the request body is empty, send a Bad Request status.
	if r.Body == nil {
		return http.StatusBadRequest, nil
	}

	var u User

	// Parses the user and checks for error.
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		return http.StatusBadRequest, nil
	}

	// The username and the password cannot be empty.
	if u.Username == "" || u.Password == "" || u.FileSystem == "" {
		return http.StatusBadRequest, errors.New("Username, password or scope are empty")
	}

	// Initialize rules if they're not initialized.
	if u.Rules == nil {
		u.Rules = []*Rule{}
	}

	// Initialize commands if not initialized.
	if u.Commands == nil {
		u.Commands = []string{}
	}

	// It's a new user so the ID will be auto created.
	if u.ID != 0 {
		u.ID = 0
	}

	// Hashes the password.
	pw, err := hashPassword(u.Password)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	u.Password = pw

	// Saves the user to the database.
	err = c.FM.db.Save(&u)
	if err == storm.ErrAlreadyExists {
		return http.StatusConflict, err
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Saves the user to the memory.
	c.FM.Users[u.Username] = &u

	// Set the Location header and return.
	w.Header().Set("Location", "/users/"+strconv.Itoa(u.ID))
	w.WriteHeader(http.StatusCreated)
	return 0, nil
}

func usersDeleteHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}

	// New users should be created on /api/users.
	if r.URL.Path == "/" {
		return http.StatusMethodNotAllowed, nil
	}

	// Otherwise we just want one, specific, user.
	sid := strings.TrimPrefix(r.URL.Path, "/")
	sid = strings.TrimSuffix(sid, "/")

	id, err := strconv.Atoi(sid)
	if err != nil {
		return http.StatusNotFound, err
	}

	err = c.FM.db.DeleteStruct(&User{ID: id})
	if err == storm.ErrNotFound {
		return http.StatusNotFound, err
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	for _, user := range c.FM.Users {
		if user.ID == id {
			delete(c.FM.Users, user.Username)
		}
	}

	return http.StatusOK, nil
}

func usersPutHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.Admin && !(r.URL.Path == "/change-password" || r.URL.Path == "/change-css") {
		return http.StatusForbidden, nil
	}

	// New users should be created on /api/users.
	if r.URL.Path == "/" {
		return http.StatusMethodNotAllowed, nil
	}

	// Otherwise we just want one, specific, user.
	sid := strings.TrimPrefix(r.URL.Path, "/")
	sid = strings.TrimSuffix(sid, "/")

	id, err := strconv.Atoi(sid)
	if err != nil && sid != "change-password" && sid != "change-css" {
		return http.StatusNotFound, err
	}

	// If the request body is empty, send a Bad Request status.
	if r.Body == nil {
		return http.StatusBadRequest, errors.New("The request has an empty body")
	}

	var u User

	// Parses the user and checks for error.
	err = json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		return http.StatusBadRequest, errors.New("Invalid JSON")
	}

	if sid == "change-password" {
		if u.Password == "" {
			return http.StatusBadRequest, errors.New("Password cannot be empty")
		}

		pw, err := hashPassword(u.Password)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		c.User.Password = pw
		err = c.FM.db.UpdateField(&User{ID: c.User.ID}, "Password", pw)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}

	if sid == "change-css" {
		c.User.CSS = u.CSS
		err = c.FM.db.UpdateField(&User{ID: c.User.ID}, "CSS", u.CSS)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}

	// The username and the filesystem cannot be empty.
	if u.Username == "" || u.FileSystem == "" {
		return http.StatusBadRequest, errors.New("Username, password or scope are empty")
	}

	// Initialize rules if they're not initialized.
	if u.Rules == nil {
		u.Rules = []*Rule{}
	}

	// Initialize commands if not initialized.
	if u.Commands == nil {
		u.Commands = []string{}
	}

	ouser, ok := c.FM.Users[u.Username]
	if !ok {
		return http.StatusNotFound, nil
	}

	u.ID = id

	if u.Password == "" {
		u.Password = ouser.Password
	} else {
		pw, err := hashPassword(u.Password)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		u.Password = pw
	}

	// Updates the whole User struct because we always are supposed
	// to send a new entire object.
	err = c.FM.db.Save(&u)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	c.FM.Users[u.Username] = &u
	return http.StatusOK, nil
}

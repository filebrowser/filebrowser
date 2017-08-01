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

type modifyRequest struct {
	What  string `json:"what"`  // Answer to: what data type?
	Which string `json:"which"` // Answer to: which field?
}

type modifyUserRequest struct {
	*modifyRequest
	Data *User `json:"data"`
}

// usersHandler is the entry point of the users API. It's just a router
// to send the request to its
func usersHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// If the user isn't admin and isn't making a PUT
	// request, then return forbidden.
	if !c.User.Admin && r.Method != http.MethodPut {
		return http.StatusForbidden, nil
	}

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

// getUserID returns the id from the user which is present
// in the request url. If the url is invalid and doesn't
// contain a valid ID, it returns an error.
func getUserID(r *http.Request) (int, error) {
	// Obtains the ID in string from the URL and converts
	// it into an integer.
	sid := strings.TrimPrefix(r.URL.Path, "/")
	sid = strings.TrimSuffix(sid, "/")
	id, err := strconv.Atoi(sid)
	if err != nil {
		return http.StatusBadRequest, err
	}

	return id, nil
}

// getUser returns the user which is present in the request
// body. If the body is empty or the JSON is invalid, it
// returns an error.
func getUser(r *http.Request) (*User, string, error) {
	// Checks if the request body is empty.
	if r.Body == nil {
		return nil, "", errEmptyRequest
	}

	// Parses the request body and checks if it's well formed.
	mod := &modifyUserRequest{}
	err := json.NewDecoder(r.Body).Decode(mod)
	if err != nil {
		return nil, "", err
	}

	// Checks if the request type is right.
	if mod.What != "user" {
		return nil, "", errWrongDataType
	}

	return mod.Data, mod.Which, nil
}

func usersGetHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// Request for the default user data.
	if r.URL.Path == "/base" {
		return renderJSON(w, c.FM.DefaultUser)
	}

	// Request for the listing of users.
	if r.URL.Path == "/" {
		users := []User{}

		for _, user := range c.FM.Users {
			// Copies the user info and removes its
			// password so it won't be sent to the
			// front-end.
			u := *user
			u.Password = ""
			users = append(users, u)
		}

		sort.Slice(users, func(i, j int) bool {
			return users[i].ID < users[j].ID
		})

		return renderJSON(w, users)
	}

	id, err := getUserID(r)
	if err != nil {
		return http.StatusInternalServerError, err
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

	// If there aren't any matches, return not found.
	return http.StatusNotFound, errUserNotExist
}

func usersPostHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path != "/" {
		return http.StatusMethodNotAllowed, nil
	}

	u, _, err := getUser(r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Checks if username isn't empty.
	if u.Username == "" {
		return http.StatusBadRequest, errEmptyUsername
	}

	// Checks if filesystem isn't empty.
	if u.FileSystem == "" {
		return http.StatusBadRequest, errEmptyScope
	}

	// Checks if password isn't empty.
	if u.Password == "" {
		return http.StatusBadRequest, errEmptyPassword
	}

	// The username, password and scope cannot be empty.
	if u.Username == "" || u.Password == "" || u.FileSystem == "" {
		return http.StatusBadRequest, errors.New("username, password or scope is empty")
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
	err = c.FM.db.Save(u)
	if err == storm.ErrAlreadyExists {
		return http.StatusConflict, errUserExist
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Saves the user to the memory.
	c.FM.Users[u.Username] = u

	// Set the Location header and return.
	w.Header().Set("Location", "/users/"+strconv.Itoa(u.ID))
	w.WriteHeader(http.StatusCreated)
	return 0, nil
}

func usersDeleteHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path == "/" {
		return http.StatusMethodNotAllowed, nil
	}

	id, err := getUserID(r)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Deletes the user from the database.
	err = c.FM.db.DeleteStruct(&User{ID: id})
	if err == storm.ErrNotFound {
		return http.StatusNotFound, errUserNotExist
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Delete the user from the in-memory users map.
	for _, user := range c.FM.Users {
		if user.ID == id {
			delete(c.FM.Users, user.Username)
			break
		}
	}

	return http.StatusOK, nil
}

func usersPutHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// New users should be created on /api/users.
	if r.URL.Path == "/" {
		return http.StatusMethodNotAllowed, nil
	}

	// Gets the user ID from the URL and checks if it's valid.
	id, err := getUserID(r)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Checks if the user has permission to access this page.
	if !c.User.Admin && id != c.User.ID {
		return http.StatusForbidden, nil
	}

	// Gets the user from the request body.
	u, which, err := getUser(r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Updates the CSS.
	if which == "css" {
		c.User.CSS = u.CSS
		err = c.FM.db.UpdateField(&User{ID: c.User.ID}, "CSS", u.CSS)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}

	// Updates the Password.
	if which == "password" {
		if u.Password == "" {
			return http.StatusBadRequest, errEmptyPassword
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

	// If can only be all.
	if which != "all" {
		return http.StatusBadRequest, errInvalidUpdateField
	}

	// Checks if username isn't empty.
	if u.Username == "" {
		return http.StatusBadRequest, errEmptyUsername
	}

	// Checks if filesystem isn't empty.
	if u.FileSystem == "" {
		return http.StatusBadRequest, errEmptyScope
	}

	// Initialize rules if they're not initialized.
	if u.Rules == nil {
		u.Rules = []*Rule{}
	}

	// Initialize commands if not initialized.
	if u.Commands == nil {
		u.Commands = []string{}
	}

	// Gets the current saved user from the in-memory map.
	var suser *User
	for _, user := range c.FM.Users {
		if user.ID == id {
			suser = user
			break
		}
	}
	if suser == nil {
		return http.StatusNotFound, nil
	}

	u.ID = id

	// Changes the password if the request wants it.
	if u.Password != "" {
		pw, err := hashPassword(u.Password)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		u.Password = pw
	} else {
		u.Password = suser.Password
	}

	// Default permissions if current are nil.
	if u.Permissions == nil {
		u.Permissions = c.FM.DefaultUser.Permissions
	}

	// Updates the whole User struct because we always are supposed
	// to send a new entire object.
	err = c.FM.db.Save(u)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// If the user changed the username, delete the old user
	// from the in-memory user map.
	if suser.Username != u.Username {
		delete(c.FM.Users, suser.Username)
	}

	c.FM.Users[u.Username] = u
	return http.StatusOK, nil
}

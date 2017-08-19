package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	fm "github.com/hacdias/filemanager"
)

type modifyRequest struct {
	What  string `json:"what"`  // Answer to: what data type?
	Which string `json:"which"` // Answer to: which field?
}

type modifyUserRequest struct {
	*modifyRequest
	Data *fm.User `json:"data"`
}

// usersHandler is the entry point of the users API. It's just a router
// to send the request to its
func usersHandler(c *fm.Context, w http.ResponseWriter, r *http.Request) (int, error) {
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
// contain a valid ID, it returns an fm.Error.
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
// returns an fm.Error.
func getUser(r *http.Request) (*fm.User, string, error) {
	// Checks if the request body is empty.
	if r.Body == nil {
		return nil, "", fm.ErrEmptyRequest
	}

	// Parses the request body and checks if it's well formed.
	mod := &modifyUserRequest{}
	err := json.NewDecoder(r.Body).Decode(mod)
	if err != nil {
		return nil, "", err
	}

	// Checks if the request type is right.
	if mod.What != "user" {
		return nil, "", fm.ErrWrongDataType
	}

	return mod.Data, mod.Which, nil
}

func usersGetHandler(c *fm.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	// Request for the default user data.
	if r.URL.Path == "/base" {
		return renderJSON(w, c.DefaultUser)
	}

	// Request for the listing of users.
	if r.URL.Path == "/" {
		users, err := c.Store.Users.Gets()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		for _, u := range users {
			// Removes the user password so it won't
			// be sent to the front-end.
			u.Password = ""
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

	u, err := c.Store.Users.Get(id)
	if err == fm.ErrExist {
		return http.StatusNotFound, err
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	u.Password = ""
	return renderJSON(w, u)
}

func usersPostHandler(c *fm.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path != "/" {
		return http.StatusMethodNotAllowed, nil
	}

	u, _, err := getUser(r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Checks if username isn't empty.
	if u.Username == "" {
		return http.StatusBadRequest, fm.ErrEmptyUsername
	}

	// Checks if filesystem isn't empty.
	if u.FileSystem == "" {
		return http.StatusBadRequest, fm.ErrEmptyScope
	}

	// Checks if password isn't empty.
	if u.Password == "" {
		return http.StatusBadRequest, fm.ErrEmptyPassword
	}

	// The username, password and scope cannot be empty.
	if u.Username == "" || u.Password == "" || u.FileSystem == "" {
		return http.StatusBadRequest, errors.New("username, password or scope is empty")
	}

	// Initialize rules if they're not initialized.
	if u.Rules == nil {
		u.Rules = []*fm.Rule{}
	}

	// Initialize commands if not initialized.
	if u.Commands == nil {
		u.Commands = []string{}
	}

	// It's a new user so the ID will be auto created.
	if u.ID != 0 {
		u.ID = 0
	}

	// Checks if the scope exists.
	if code, err := checkFS(string(u.FileSystem)); err != nil {
		return code, err
	}

	// Hashes the password.
	pw, err := fm.HashPassword(u.Password)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	u.Password = pw

	// Saves the user to the database.
	err = c.Store.Users.Save(u)
	if err == fm.ErrExist {
		return http.StatusConflict, err
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Set the Location header and return.
	w.Header().Set("Location", "/users/"+strconv.Itoa(u.ID))
	w.WriteHeader(http.StatusCreated)
	return 0, nil
}

func checkFS(path string) (int, error) {
	info, err := os.Stat(path)

	if err != nil {
		if !os.IsNotExist(err) {
			return http.StatusInternalServerError, err
		}

		err = os.MkdirAll(path, 0666)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return 0, nil
	}

	if !info.IsDir() {
		return http.StatusBadRequest, errors.New("Scope is not a dir")
	}

	return 0, nil
}

func usersDeleteHandler(c *fm.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path == "/" {
		return http.StatusMethodNotAllowed, nil
	}

	id, err := getUserID(r)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Deletes the user from the database.
	err = c.Store.Users.Delete(id)
	if err == fm.ErrNotExist {
		return http.StatusNotFound, fm.ErrNotExist
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func usersPutHandler(c *fm.Context, w http.ResponseWriter, r *http.Request) (int, error) {
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

	// Updates the CSS and locale.
	if which == "partial" {
		c.User.CSS = u.CSS
		c.User.Locale = u.Locale

		err = c.Store.Users.Update(c.User, "CSS", "Locale")
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}

	// Updates the Password.
	if which == "password" {
		if u.Password == "" {
			return http.StatusBadRequest, fm.ErrEmptyPassword
		}

		c.User.Password, err = fm.HashPassword(u.Password)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		err = c.Store.Users.Update(c.User, "Password")
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}

	// If can only be all.
	if which != "all" {
		return http.StatusBadRequest, fm.ErrInvalidUpdateField
	}

	// Checks if username isn't empty.
	if u.Username == "" {
		return http.StatusBadRequest, fm.ErrEmptyUsername
	}

	// Checks if filesystem isn't empty.
	if u.FileSystem == "" {
		return http.StatusBadRequest, fm.ErrEmptyScope
	}

	// Checks if the scope exists.
	if code, err := checkFS(string(u.FileSystem)); err != nil {
		return code, err
	}

	// Initialize rules if they're not initialized.
	if u.Rules == nil {
		u.Rules = []*fm.Rule{}
	}

	// Initialize commands if not initialized.
	if u.Commands == nil {
		u.Commands = []string{}
	}

	// Gets the current saved user from the in-memory map.
	suser, err := c.Store.Users.Get(id)
	if err == fm.ErrNotExist {
		return http.StatusNotFound, nil
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	u.ID = id

	// Changes the password if the request wants it.
	if u.Password != "" {
		pw, err := fm.HashPassword(u.Password)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		u.Password = pw
	} else {
		u.Password = suser.Password
	}

	// Updates the whole User struct because we always are supposed
	// to send a new entire object.
	err = c.Store.Users.Update(u)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

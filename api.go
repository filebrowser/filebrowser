package filemanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/asdine/storm"
)

func cleanURL(path string) (string, string) {
	if path == "" {
		return "", ""
	}

	path = strings.TrimPrefix(path, "/")

	i := strings.Index(path, "/")
	if i == -1 {
		return "", path
	}

	return path[0:i], path[i:len(path)]
}

func serveAPI(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path == "/auth/get" {
		return authHandler(c, w, r)
	}

	if r.URL.Path == "/auth/renew" {
		return renewAuthHandler(c, w, r)
	}

	valid, _ := validateAuth(c, r)
	if !valid {
		return http.StatusForbidden, nil
	}

	var router string
	router, r.URL.Path = cleanURL(r.URL.Path)

	if !c.us.Allowed(r.URL.Path) {
		return http.StatusForbidden, nil
	}

	if router == "checksum" || router == "download" {
		var err error
		c.fi, err = getInfo(r.URL, c.fm, c.us)
		if err != nil {
			return errorToHTTP(err, false), err
		}
	}

	fmt.Println(c.us)

	switch router {
	case "download":
		return downloadHandler(c, w, r)
	case "checksum":
		return checksumHandler(c, w, r)
	case "command":
		return command(c, w, r)
	case "search":
		return search(c, w, r)
	case "resource":
		return resourceHandler(c, w, r)
	case "users":
		if !c.us.Admin && !(r.URL.Path == "/self" && r.Method == http.MethodPut) {
			return http.StatusForbidden, nil
		}

		return usersHandler(c, w, r)
	}

	return http.StatusNotFound, nil
}

func resourceHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	switch r.Method {
	case http.MethodGet:
		return getHandler(c, w, r)
	case http.MethodDelete:
		return deleteHandler(c, w, r)
	case http.MethodPut:
		return postPutHandler(c, w, r)
	case http.MethodPatch:
		return patchHandler(c, w, r)
	case http.MethodPost:
		return postPutHandler(c, w, r)
	}

	/* // Execute beforeSave if it is a PUT request.
	if r.Method == http.MethodPut {
		if err := c.fm.BeforeSave(r, c.fm, c.us); err != nil {
			return http.StatusInternalServerError, err
		}
	}

	// Execute afterSave if it is a PUT request.
	if r.Method == http.MethodPut {
		if err := c.fm.AfterSave(r, c.fm, c.us); err != nil {
			return http.StatusInternalServerError, err
		}
	} */

	return http.StatusNotImplemented, nil
}

func getHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// Obtains the information of the directory/file.
	f, err := getInfo(r.URL, c.fm, c.us)
	if err != nil {
		return errorToHTTP(err, false), err
	}

	// If it's a dir and the path doesn't end with a trailing slash,
	// redirect the user.
	if f.IsDir && !strings.HasSuffix(r.URL.Path, "/") {
		r.URL.Path = r.URL.Path + "/"
	}

	// If it is a dir, go and serve the listing.
	if f.IsDir {
		c.fi = f
		return listingHandler(c, w, r)
	}

	// Tries to get the file type.
	if err = f.RetrieveFileType(); err != nil {
		return errorToHTTP(err, true), err
	}

	// If it can't be edited or the user isn't allowed to,
	// serve it as a listing, with a preview of the file.
	if !f.CanBeEdited() || !c.us.AllowEdit {
		f.Kind = "preview"
	} else {
		// Otherwise, we just bring the editor in!
		f.Kind = "editor"

		err = f.getEditor()
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	return renderJSON(w, f)
}

func listingHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	f := c.fi
	f.Kind = "listing"

	err := f.getListing(c, r)
	if err != nil {
		return errorToHTTP(err, true), err
	}

	listing := f.listing

	cookieScope := c.fm.RootURL()
	if cookieScope == "" {
		cookieScope = "/"
	}

	// Copy the query values into the Listing struct
	listing.Sort, listing.Order, err = handleSortOrder(w, r, cookieScope)
	if err != nil {
		return http.StatusBadRequest, err
	}

	listing.ApplySort()
	listing.Display = displayMode(w, r, cookieScope)

	return renderJSON(w, f)
}

func deleteHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// Prevent the removal of the root directory.
	if r.URL.Path == "/" {
		return http.StatusForbidden, nil
	}

	// Remove the file or folder.
	err := c.us.FileSystem.RemoveAll(context.TODO(), r.URL.Path)
	if err != nil {
		return errorToHTTP(err, true), err
	}

	return http.StatusOK, nil
}

func postPutHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// Checks if the current request is for a directory and not a file.
	if strings.HasSuffix(r.URL.Path, "/") {
		// If the method is PUT, we return 405 Method not Allowed, because
		// POST should be used instead.
		if r.Method == http.MethodPut {
			return http.StatusMethodNotAllowed, nil
		}

		// Otherwise we try to create the directory.
		err := c.us.FileSystem.Mkdir(context.TODO(), r.URL.Path, 0666)
		return errorToHTTP(err, false), err
	}

	// If using POST method, we are trying to create a new file so it is not
	// desirable to override an already existent file. Thus, we check
	// if the file already exists. If so, we just return a 409 Conflict.
	if r.Method == http.MethodPost {
		if _, err := c.us.FileSystem.Stat(context.TODO(), r.URL.Path); err == nil {
			return http.StatusConflict, nil
		}
	}

	// Create/Open the file.
	f, err := c.us.FileSystem.OpenFile(context.TODO(), r.URL.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer f.Close()

	if err != nil {
		return errorToHTTP(err, false), err
	}

	// Copies the new content for the file.
	_, err = io.Copy(f, r.Body)
	if err != nil {
		return errorToHTTP(err, false), err
	}

	// Gets the info about the file.
	fi, err := f.Stat()
	if err != nil {
		return errorToHTTP(err, false), err
	}

	// Writes the ETag Header.
	etag := fmt.Sprintf(`"%x%x"`, fi.ModTime().UnixNano(), fi.Size())
	w.Header().Set("ETag", etag)
	return http.StatusOK, nil
}

func patchHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	dst := r.Header.Get("Destination")
	dst, err := url.QueryUnescape(dst)
	if err != nil {
		return errorToHTTP(err, true), err
	}

	src := r.URL.Path

	if dst == "/" || src == "/" {
		return http.StatusForbidden, nil
	}

	err = c.us.FileSystem.Rename(context.TODO(), src, dst)
	return errorToHTTP(err, true), err
}

// displayMode obtaisn the display mode from URL, or from the
// cookie.
func displayMode(w http.ResponseWriter, r *http.Request, scope string) string {
	displayMode := r.URL.Query().Get("display")

	if displayMode == "" {
		if displayCookie, err := r.Cookie("display"); err == nil {
			displayMode = displayCookie.Value
		}
	}

	if displayMode == "" || (displayMode != "mosaic" && displayMode != "list") {
		displayMode = "mosaic"
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "display",
		Value:  displayMode,
		MaxAge: 31536000,
		Path:   scope,
		Secure: r.TLS != nil,
	})

	return displayMode
}

// handleSortOrder gets and stores for a Listing the 'sort' and 'order',
// and reads 'limit' if given. The latter is 0 if not given. Sets cookies.
func handleSortOrder(w http.ResponseWriter, r *http.Request, scope string) (sort string, order string, err error) {
	sort = r.URL.Query().Get("sort")
	order = r.URL.Query().Get("order")

	// If the query 'sort' or 'order' is empty, use defaults or any values
	// previously saved in Cookies.
	switch sort {
	case "":
		sort = "name"
		if sortCookie, sortErr := r.Cookie("sort"); sortErr == nil {
			sort = sortCookie.Value
		}
	case "name", "size":
		http.SetCookie(w, &http.Cookie{
			Name:   "sort",
			Value:  sort,
			MaxAge: 31536000,
			Path:   scope,
			Secure: r.TLS != nil,
		})
	}

	switch order {
	case "":
		order = "asc"
		if orderCookie, orderErr := r.Cookie("order"); orderErr == nil {
			order = orderCookie.Value
		}
	case "asc", "desc":
		http.SetCookie(w, &http.Cookie{
			Name:   "order",
			Value:  order,
			MaxAge: 31536000,
			Path:   scope,
			Secure: r.TLS != nil,
		})
	}

	return
}

func usersHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
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
func usersGetHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// If the request is a list of users.
	if r.URL.Path == "/" {
		users := []User{}

		for _, user := range c.fm.Users {
			// Copies the user and removes the password.
			u := *user
			u.Password = ""
			users = append(users, u)
		}

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
	for _, user := range c.fm.Users {
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

func usersPostHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
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
	err = c.fm.db.Save(&u)
	if err == storm.ErrAlreadyExists {
		return http.StatusConflict, err
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Saves the user to the memory.
	c.fm.Users[u.Username] = &u

	// Set the Location header and return.
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Location", c.fm.RootURL()+"/api/users/"+strconv.Itoa(u.ID))
	return 0, nil
}

func usersDeleteHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
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

	err = c.fm.db.DeleteStruct(&User{ID: id})
	if err == storm.ErrNotFound {
		return http.StatusNotFound, err
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	for _, user := range c.fm.Users {
		if user.ID == id {
			delete(c.fm.Users, user.Username)
		}
	}

	return http.StatusOK, nil
}

func usersPutHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// New users should be created on /api/users.
	if r.URL.Path == "/" {
		return http.StatusMethodNotAllowed, nil
	}

	// Otherwise we just want one, specific, user.
	sid := strings.TrimPrefix(r.URL.Path, "/")
	sid = strings.TrimSuffix(sid, "/")

	id, err := strconv.Atoi(sid)
	if err != nil && sid != "self" {
		return http.StatusNotFound, err
	}

	// If the request body is empty, send a Bad Request status.
	if r.Body == nil {
		return http.StatusBadRequest, nil
	}

	var u User

	// Parses the user and checks for error.
	err = json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		return http.StatusBadRequest, nil
	}

	if sid == "self" {
		if u.Password == "" {
			return http.StatusBadRequest, errors.New("Password missing")
		}

		pw, err := hashPassword(u.Password)
		if err != nil {
			fmt.Println(err)
			return http.StatusInternalServerError, err
		}

		c.us.Password = pw
		err = c.fm.db.UpdateField(&User{ID: c.us.ID}, "Password", pw)
		if err != nil {
			fmt.Println(err)
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

	ouser, ok := c.fm.Users[u.Username]
	if !ok {
		return http.StatusNotFound, nil
	}

	u.ID = id
	u.Password = ouser.Password

	// Updates the whole User struct because we always are supposed
	// to send a new entire object.
	err = c.fm.db.Save(&u)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	c.fm.Users[u.Username] = &u
	return http.StatusOK, nil
}

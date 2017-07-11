package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func resourceHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	switch r.Method {
	case http.MethodGet:
		return resourceGetHandler(c, w, r)
	case http.MethodDelete:
		return resourceDeleteHandler(c, w, r)
	case http.MethodPut:
		// Before save command handler.
		path := filepath.Join(string(c.User.FileSystem), r.URL.Path)
		if err := c.FM.Runner("before_save", path); err != nil {
			return http.StatusInternalServerError, err
		}

		code, err := resourcePostPutHandler(c, w, r)
		if code != http.StatusOK {
			return code, err
		}

		// After save command handler.
		if err := c.FM.Runner("after_save", path); err != nil {
			return http.StatusInternalServerError, err
		}

		return code, err
	case http.MethodPatch:
		return resourcePatchHandler(c, w, r)
	case http.MethodPost:
		return resourcePostPutHandler(c, w, r)
	}

	return http.StatusNotImplemented, nil
}

func resourceGetHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// Obtains the information of the directory/file.
	f, err := getInfo(r.URL, c.FM, c.User)
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
		c.FI = f
		return listingHandler(c, w, r)
	}

	// Tries to get the file type.
	if err = f.RetrieveFileType(true); err != nil {
		return errorToHTTP(err, true), err
	}

	// If it can't be edited or the user isn't allowed to,
	// serve it as a listing, with a preview of the file.
	if !f.CanBeEdited() || !c.User.AllowEdit {
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

func listingHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	f := c.FI
	f.Kind = "listing"

	err := f.getListing(c, r)
	if err != nil {
		return errorToHTTP(err, true), err
	}

	listing := f.listing

	cookieScope := c.FM.RootURL()
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

func resourceDeleteHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// Prevent the removal of the root directory.
	if r.URL.Path == "/" {
		return http.StatusForbidden, nil
	}

	// Remove the file or folder.
	err := c.User.FileSystem.RemoveAll(context.TODO(), r.URL.Path)
	if err != nil {
		return errorToHTTP(err, true), err
	}

	return http.StatusOK, nil
}

func resourcePostPutHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// Checks if the current request is for a directory and not a file.
	if strings.HasSuffix(r.URL.Path, "/") {
		// If the method is PUT, we return 405 Method not Allowed, because
		// POST should be used instead.
		if r.Method == http.MethodPut {
			return http.StatusMethodNotAllowed, nil
		}

		// Otherwise we try to create the directory.
		err := c.User.FileSystem.Mkdir(context.TODO(), r.URL.Path, 0666)
		return errorToHTTP(err, false), err
	}

	// If using POST method, we are trying to create a new file so it is not
	// desirable to override an already existent file. Thus, we check
	// if the file already exists. If so, we just return a 409 Conflict.
	if r.Method == http.MethodPost {
		if _, err := c.User.FileSystem.Stat(context.TODO(), r.URL.Path); err == nil {
			return http.StatusConflict, errors.New("There is already a file on that path")
		}
	}

	// Create/Open the file.
	f, err := c.User.FileSystem.OpenFile(context.TODO(), r.URL.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
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

func resourcePatchHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	dst := r.Header.Get("Destination")
	dst, err := url.QueryUnescape(dst)
	if err != nil {
		return errorToHTTP(err, true), err
	}

	src := r.URL.Path

	if dst == "/" || src == "/" {
		return http.StatusForbidden, nil
	}

	err = c.User.FileSystem.Rename(context.TODO(), src, dst)
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

package http

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	fb "github.com/filebrowser/filebrowser/lib"
	"github.com/hacdias/fileutils"
)

// sanitizeURL sanitizes the URL to prevent path transversal
// using fileutils.SlashClean and adds the trailing slash bar.
func sanitizeURL(url string) string {
	path := fileutils.SlashClean(url)
	if strings.HasSuffix(url, "/") && path != "/" {
		return path + "/"
	}
	return path
}

func resourceHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	r.URL.Path = sanitizeURL(r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		return resourceGetHandler(c, w, r)
	case http.MethodDelete:
		return resourceDeleteHandler(c, w, r)
	case http.MethodPut:
		// Before save command handler.
		path := filepath.Join(c.User.Scope, r.URL.Path)
		if err := c.Runner("before_save", path, "", c.User); err != nil {
			return http.StatusInternalServerError, err
		}

		code, err := resourcePostPutHandler(c, w, r)
		if code != http.StatusOK {
			return code, err
		}

		// After save command handler.
		if err := c.Runner("after_save", path, "", c.User); err != nil {
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

func resourceGetHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	// Gets the information of the directory/file.
	f, err := fb.GetInfo(r.URL, c.FileBrowser, c.User)
	if err != nil {
		return ErrorToHTTP(err, false), err
	}

	// If it's a dir and the path doesn't end with a trailing slash,
	// add a trailing slash to the path.
	if f.IsDir && !strings.HasSuffix(r.URL.Path, "/") {
		r.URL.Path = r.URL.Path + "/"
	}

	// If it is a dir, go and serve the listing.
	if f.IsDir {
		c.File = f
		return listingHandler(c, w, r)
	}

	// Tries to get the file type.
	if err = f.GetFileType(true); err != nil {
		return ErrorToHTTP(err, true), err
	}

	// Serve a preview if the file can't be edited or the
	// user has no permission to edit this file. Otherwise,
	// just serve the editor.
	if !f.CanBeEdited() || !c.User.AllowEdit {
		f.Kind = "preview"
		return renderJSON(w, f)
	}

	f.Kind = "editor"

	// Tries to get the editor data.
	if err = f.GetEditor(); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, f)
}

func listingHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	f := c.File
	f.Kind = "listing"

	// Tries to get the listing data.
	if err := f.GetListing(c.User, r); err != nil {
		return ErrorToHTTP(err, true), err
	}

	listing := f.Listing

	// Defines the cookie scope.
	cookieScope := c.RootURL()
	if cookieScope == "" {
		cookieScope = "/"
	}

	// Copy the query values into the Listing struct
	if sort, order, err := handleSortOrder(w, r, cookieScope); err == nil {
		listing.Sort = sort
		listing.Order = order
	} else {
		return http.StatusBadRequest, err
	}

	listing.ApplySort()
	return renderJSON(w, f)
}

func resourceDeleteHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	// Prevent the removal of the root directory.
	if r.URL.Path == "/" || !c.User.AllowEdit {
		return http.StatusForbidden, nil
	}

	// Fire the before trigger.
	if err := c.Runner("before_delete", r.URL.Path, "", c.User); err != nil {
		return http.StatusInternalServerError, err
	}

	// Remove the file or folder.
	err := c.User.FileSystem.RemoveAll(r.URL.Path)
	if err != nil {
		return ErrorToHTTP(err, true), err
	}

	// Fire the after trigger.
	if err := c.Runner("after_delete", r.URL.Path, "", c.User); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func resourcePostPutHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.AllowNew && r.Method == http.MethodPost {
		return http.StatusForbidden, nil
	}

	if !c.User.AllowEdit && r.Method == http.MethodPut {
		return http.StatusForbidden, nil
	}

	// Discard any invalid upload before returning to avoid connection
	// reset error.
	defer func() {
		io.Copy(ioutil.Discard, r.Body)
	}()

	// Checks if the current request is for a directory and not a file.
	if strings.HasSuffix(r.URL.Path, "/") {
		// If the method is PUT, we return 405 Method not Allowed, because
		// POST should be used instead.
		if r.Method == http.MethodPut {
			return http.StatusMethodNotAllowed, nil
		}

		// Otherwise we try to create the directory.
		err := c.User.FileSystem.Mkdir(r.URL.Path, 0775)
		return ErrorToHTTP(err, false), err
	}

	// If using POST method, we are trying to create a new file so it is not
	// desirable to override an already existent file. Thus, we check
	// if the file already exists. If so, we just return a 409 Conflict.
	if r.Method == http.MethodPost && r.Header.Get("Action") != "override" {
		if _, err := c.User.FileSystem.Stat(r.URL.Path); err == nil {
			return http.StatusConflict, errors.New("There is already a file on that path")
		}
	}

	// Fire the before trigger.
	if err := c.Runner("before_upload", r.URL.Path, "", c.User); err != nil {
		return http.StatusInternalServerError, err
	}

	// Create/Open the file.
	f, err := c.User.FileSystem.OpenFile(r.URL.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
	if err != nil {
		return ErrorToHTTP(err, false), err
	}
	defer f.Close()

	// Copies the new content for the file.
	_, err = io.Copy(f, r.Body)
	if err != nil {
		return ErrorToHTTP(err, false), err
	}

	// Gets the info about the file.
	fi, err := f.Stat()
	if err != nil {
		return ErrorToHTTP(err, false), err
	}

	// Check if this instance has a Static Generator and handles publishing
	// or scheduling if it's the case.
	if c.StaticGen != nil {
		code, err := resourcePublishSchedule(c, w, r)
		if code != 0 {
			return code, err
		}
	}

	// Writes the ETag Header.
	etag := fmt.Sprintf(`"%x%x"`, fi.ModTime().UnixNano(), fi.Size())
	w.Header().Set("ETag", etag)

	// Fire the after trigger.
	if err := c.Runner("after_upload", r.URL.Path, "", c.User); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func resourcePublishSchedule(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	publish := r.Header.Get("Publish")
	schedule := r.Header.Get("Schedule")

	if publish != "true" && schedule == "" {
		return 0, nil
	}

	if !c.User.AllowPublish {
		return http.StatusForbidden, nil
	}

	if publish == "true" {
		return resourcePublish(c, w, r)
	}

	t, err := time.Parse("2006-01-02T15:04", schedule)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	c.Cron.AddFunc(t.Format("05 04 15 02 01 *"), func() {
		_, err := resourcePublish(c, w, r)
		if err != nil {
			log.Print(err)
		}
	})

	return http.StatusOK, nil
}

func resourcePublish(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	path := filepath.Join(c.User.Scope, r.URL.Path)

	// Before save command handler.
	if err := c.Runner("before_publish", path, "", c.User); err != nil {
		return http.StatusInternalServerError, err
	}

	code, err := c.StaticGen.Publish(c, w, r)
	if err != nil {
		return code, err
	}

	// Executed the before publish command.
	if err := c.Runner("before_publish", path, "", c.User); err != nil {
		return http.StatusInternalServerError, err
	}

	return code, nil
}

// resourcePatchHandler is the entry point for resource handler.
func resourcePatchHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.AllowEdit {
		return http.StatusForbidden, nil
	}

	dst := r.Header.Get("Destination")
	action := r.Header.Get("Action")
	dst, err := url.QueryUnescape(dst)
	if err != nil {
		return ErrorToHTTP(err, true), err
	}

	src := r.URL.Path

	if dst == "/" || src == "/" {
		return http.StatusForbidden, nil
	}

	if action == "copy" {
		// Fire the after trigger.
		if err := c.Runner("before_copy", src, dst, c.User); err != nil {
			return http.StatusInternalServerError, err
		}

		// Copy the file.
		err = c.User.FileSystem.Copy(src, dst)

		// Fire the after trigger.
		if err := c.Runner("after_copy", src, dst, c.User); err != nil {
			return http.StatusInternalServerError, err
		}
	} else {
		// Fire the after trigger.
		if err := c.Runner("before_rename", src, dst, c.User); err != nil {
			return http.StatusInternalServerError, err
		}

		// Rename the file.
		err = c.User.FileSystem.Rename(src, dst)

		// Fire the after trigger.
		if err := c.Runner("after_rename", src, dst, c.User); err != nil {
			return http.StatusInternalServerError, err
		}
	}

	return ErrorToHTTP(err, true), err
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

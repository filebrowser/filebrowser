package filemanager

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mholt/archiver"
)

// assetsURL is the url where static assets are served.
const assetsURL = "/_internal"

func serveHTTP(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var (
		code int
		err  error
	)

	// Checks if the URL contains the baseURL. If so, it strips it. Otherwise,
	// it throws an error.
	if p := strings.TrimPrefix(r.URL.Path, c.fm.baseURL); len(p) < len(r.URL.Path) {
		r.URL.Path = p
	} else {
		return http.StatusNotFound, nil
	}

	// Checks if the URL matches the Assets URL. Returns the asset if the
	// method is GET and Status Forbidden otherwise.
	if matchURL(r.URL.Path, assetsURL) {
		if r.Method == http.MethodGet {
			return serveAssets(c, w, r)
		}

		return http.StatusForbidden, nil
	}

	username, _, _ := r.BasicAuth()
	if _, ok := c.fm.Users[username]; ok {
		c.us = c.fm.Users[username]
	} else {
		c.us = c.fm.User
	}

	// Checks if the request URL is for the WebDav server.
	if matchURL(r.URL.Path, c.fm.webDavURL) {
		return serveWebDAV(c, w, r)
	}

	w.Header().Set("x-frame-options", "SAMEORIGIN")
	w.Header().Set("x-content-type", "nosniff")
	w.Header().Set("x-xss-protection", "1; mode=block")

	// Checks if the User is allowed to access this file
	if !c.us.Allowed(r.URL.Path) {
		if r.Method == http.MethodGet {
			return htmlError(
				w, http.StatusForbidden,
				errors.New("You don't have permission to access this page"),
			)
		}

		return http.StatusForbidden, nil
	}

	if r.URL.Query().Get("search") != "" {
		return search(c, w, r)
	}

	if r.URL.Query().Get("command") != "" {
		return command(c, w, r)
	}

	if r.Method == http.MethodGet {
		var f *fileInfo

		// Obtains the information of the directory/file.
		f, err = getInfo(r.URL, c.fm, c.us)
		if err != nil {
			if r.Method == http.MethodGet {
				return htmlError(w, code, err)
			}

			code = errorToHTTP(err, false)
			return code, err
		}

		c.fi = f

		// If it's a dir and the path doesn't end with a trailing slash,
		// redirect the user.
		if f.IsDir && !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, c.fm.RootURL()+r.URL.Path+"/", http.StatusTemporaryRedirect)
			return 0, nil
		}

		switch {
		case r.URL.Query().Get("download") != "":
			code, err = serveDownload(c, w, r)
		case !f.IsDir && r.URL.Query().Get("checksum") != "":
			code, err = serveChecksum(c, w, r)
		case r.URL.Query().Get("raw") == "true" && !f.IsDir:
			http.ServeFile(w, r, f.Path)
			code, err = 0, nil
		case f.IsDir:
			code, err = serveListing(c, w, r)
		default:
			code, err = serveSingle(c, w, r)
		}

		if err != nil {
			code, err = htmlError(w, code, err)
		}

		return code, err
	}

	return http.StatusNotImplemented, nil
}

// serveWebDAV handles the webDAV route of the File Manager.
func serveWebDAV(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var err error

	// Checks for user permissions relatively to this path.
	if !c.us.Allowed(strings.TrimPrefix(r.URL.Path, c.fm.webDavURL)) {
		return http.StatusForbidden, nil
	}

	switch r.Method {
	case "GET", "HEAD":
		// Excerpt from RFC4918, section 9.4:
		//
		// 		GET, when applied to a collection, may return the contents of an
		//		"index.html" resource, a human-readable view of the contents of
		//		the collection, or something else altogether.
		//
		// It was decided on https://github.com/hacdias/caddy-filemanager/issues/85
		// that GET, for collections, will return the same as PROPFIND method.
		path := strings.Replace(r.URL.Path, c.fm.webDavURL, "", 1)
		path = c.us.scope + "/" + path
		path = filepath.Clean(path)

		var i os.FileInfo
		i, err = os.Stat(path)
		if err != nil {
			// Is there any error? WebDav will handle it... no worries.
			break
		}

		if i.IsDir() {
			r.Method = "PROPFIND"

			if r.Method == "HEAD" {
				w = newResponseWriterNoBody(w)
			}
		}
	case "PROPPATCH", "MOVE", "PATCH", "PUT", "DELETE":
		if !c.us.AllowEdit {
			return http.StatusForbidden, nil
		}
	case "MKCOL", "COPY":
		if !c.us.AllowNew {
			return http.StatusForbidden, nil
		}
	}

	// Preprocess the PUT request if it's the case
	if r.Method == http.MethodPut {
		if err = c.fm.BeforeSave(r, c.fm, c.us); err != nil {
			return http.StatusInternalServerError, err
		}

		if put(c, w, r) != nil {
			return http.StatusInternalServerError, err
		}
	}

	c.fm.handler.ServeHTTP(w, r)
	if err = c.fm.AfterSave(r, c.fm, c.us); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

// Serve provides the needed assets for the front-end
func serveAssets(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// gets the filename to be used with Assets function
	filename := strings.TrimPrefix(r.URL.Path, assetsURL)

	var file []byte
	var err error

	switch {
	case strings.HasPrefix(filename, "/css"):
		filename = strings.Replace(filename, "/css/", "", 1)
		file, err = c.fm.assets.css.Bytes(filename)
	case strings.HasPrefix(filename, "/js"):
		filename = strings.Replace(filename, "/js/", "", 1)
		file, err = c.fm.assets.js.Bytes(filename)
	default:
		err = errors.New("not found")
	}

	if err != nil {
		return http.StatusNotFound, nil
	}

	// Get the file extension and its mimetype
	extension := filepath.Ext(filename)
	mediatype := mime.TypeByExtension(extension)

	// Write the header with the Content-Type and write the file
	// content to the buffer
	w.Header().Set("Content-Type", mediatype)
	w.Write(file)
	return 200, nil
}

// serveChecksum calculates the hash of a file. Supports MD5, SHA1, SHA256 and SHA512.
func serveChecksum(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	query := r.URL.Query().Get("checksum")

	val, err := c.fi.Checksum(query)
	if err == errInvalidOption {
		return http.StatusBadRequest, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Write([]byte(val))
	return http.StatusOK, nil
}

// serveSingle serves a single file in an editor (if it is editable), shows the
// plain file, or downloads it if it can't be shown.
func serveSingle(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var err error

	if err = c.fi.RetrieveFileType(); err != nil {
		return errorToHTTP(err, true), err
	}

	p := &page{
		Name:      c.fi.Name,
		Path:      c.fi.VirtualPath,
		IsDir:     false,
		Data:      c.fi,
		User:      c.us,
		PrefixURL: c.fm.prefixURL,
		BaseURL:   c.fm.RootURL(),
		WebDavURL: c.fm.WebDavURL(),
	}

	// If the request accepts JSON, we send the file information.
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		return p.PrintAsJSON(w)
	}

	if c.fi.Type == "text" {
		if err = c.fi.Read(); err != nil {
			return errorToHTTP(err, true), err
		}
	}

	if c.fi.CanBeEdited() && c.us.AllowEdit {
		p.Data, err = getEditor(r, c.fi)
		p.Editor = true
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return p.PrintAsHTML(w, c.fm.assets.templates, "frontmatter", "editor")
	}

	return p.PrintAsHTML(w, c.fm.assets.templates, "single")
}

// serveDownload creates an archive in one of the supported formats (zip, tar,
// tar.gz or tar.bz2) and sends it to be downloaded.
func serveDownload(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	query := r.URL.Query().Get("download")

	if !c.fi.IsDir {
		w.Header().Set("Content-Disposition", "attachment; filename="+c.fi.Name)
		http.ServeFile(w, r, c.fi.Path)
		return 0, nil
	}

	files := []string{}
	names := strings.Split(r.URL.Query().Get("files"), ",")

	if len(names) != 0 {
		for _, name := range names {
			name, err := url.QueryUnescape(name)

			if err != nil {
				return http.StatusInternalServerError, err
			}

			files = append(files, filepath.Join(c.fi.Path, name))
		}

	} else {
		files = append(files, c.fi.Path)
	}

	if query == "true" {
		query = "zip"
	}

	var (
		extension string
		temp      string
		err       error
		tempfile  string
	)

	temp, err = ioutil.TempDir("", "")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	defer os.RemoveAll(temp)
	tempfile = filepath.Join(temp, "temp")

	switch query {
	case "zip":
		extension, err = ".zip", archiver.Zip.Make(tempfile, files)
	case "tar":
		extension, err = ".tar", archiver.Tar.Make(tempfile, files)
	case "targz":
		extension, err = ".tar.gz", archiver.TarGz.Make(tempfile, files)
	case "tarbz2":
		extension, err = ".tar.bz2", archiver.TarBz2.Make(tempfile, files)
	case "tarxz":
		extension, err = ".tar.xz", archiver.TarXZ.Make(tempfile, files)
	default:
		return http.StatusNotImplemented, nil
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	file, err := os.Open(temp + "/temp")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	name := c.fi.Name
	if name == "." || name == "" {
		name = "download"
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+name+extension)
	io.Copy(w, file)
	return http.StatusOK, nil
}

// serveListing presents the user with a listage of a directory folder.
func serveListing(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var err error

	// Loads the content of the directory.
	listing, err := getListing(c.us, c.fi.VirtualPath, c.fm.RootURL()+r.URL.Path)
	if err != nil {
		return errorToHTTP(err, true), err
	}

	cookieScope := c.fm.RootURL()
	if cookieScope == "" {
		cookieScope = "/"
	}

	// Copy the query values into the Listing struct
	var limit int
	listing.Sort, listing.Order, limit, err = handleSortOrder(w, r, cookieScope)
	if err != nil {
		return http.StatusBadRequest, err
	}

	listing.ApplySort()

	if limit > 0 && limit <= len(listing.Items) {
		listing.Items = listing.Items[:limit]
		listing.ItemsLimitedTo = limit
	}

	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		marsh, err := json.Marshal(listing.Items)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if _, err := w.Write(marsh); err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}

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
		Path:   cookieScope,
		Secure: r.TLS != nil,
	})

	p := &page{
		minimal:   r.Header.Get("Minimal") == "true",
		Name:      listing.Name,
		Path:      c.fi.VirtualPath,
		IsDir:     true,
		User:      c.us,
		PrefixURL: c.fm.prefixURL,
		BaseURL:   c.fm.RootURL(),
		WebDavURL: c.fm.WebDavURL(),
		Display:   displayMode,
		Data:      listing,
	}

	return p.PrintAsHTML(w, c.fm.assets.templates, "listing")
}

// handleSortOrder gets and stores for a Listing the 'sort' and 'order',
// and reads 'limit' if given. The latter is 0 if not given. Sets cookies.
func handleSortOrder(w http.ResponseWriter, r *http.Request, scope string) (sort string, order string, limit int, err error) {
	sort = r.URL.Query().Get("sort")
	order = r.URL.Query().Get("order")
	limitQuery := r.URL.Query().Get("limit")

	// If the query 'sort' or 'order' is empty, use defaults or any values
	// previously saved in Cookies.
	switch sort {
	case "":
		sort = "name"
		if sortCookie, sortErr := r.Cookie("sort"); sortErr == nil {
			sort = sortCookie.Value
		}
	case "name", "size", "type":
		http.SetCookie(w, &http.Cookie{
			Name:   "sort",
			Value:  sort,
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
			Path:   scope,
			Secure: r.TLS != nil,
		})
	}

	if limitQuery != "" {
		limit, err = strconv.Atoi(limitQuery)
		// If the 'limit' query can't be interpreted as a number, return err.
		if err != nil {
			return
		}
	}

	return
}

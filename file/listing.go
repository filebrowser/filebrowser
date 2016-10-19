package file

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/page"
	"github.com/hacdias/caddy-filemanager/utils"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// A Listing is the context used to fill out a template.
type Listing struct {
	// The name of the directory (the last element of the path)
	Name string
	// The full path of the request
	Path string
	// The items (files and folders) in the path
	Items []Info
	// The number of directories in the listing
	NumDirs int
	// The number of files (items that aren't directories) in the listing
	NumFiles int
	// Which sorting order is used
	Sort string
	// And which order
	Order string
	// If ≠0 then Items have been limited to that many elements
	ItemsLimitedTo     int
	httpserver.Context `json:"-"`
}

func (i *Info) serveListing(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.User) (int, error) {
	var err error

	file, err := u.FileSystem.OpenFile(i.VirtualPath, os.O_RDONLY, 0)
	if err != nil {
		return utils.ErrorToHTTPCode(err, true), err
	}
	defer file.Close()

	listing, err := i.loadDirectoryContents(file, r.URL.Path, u)
	if err != nil {
		fmt.Println(err)
		switch {
		case os.IsPermission(err):
			return http.StatusForbidden, err
		case os.IsExist(err):
			return http.StatusGone, err
		default:
			return http.StatusInternalServerError, err
		}
	}

	listing.Context = httpserver.Context{
		Root: http.Dir(u.Scope),
		Req:  r,
		URL:  r.URL,
	}

	// Copy the query values into the Listing struct
	var limit int
	listing.Sort, listing.Order, limit, err = handleSortOrder(w, r, c.Scope)
	if err != nil {
		return http.StatusBadRequest, err
	}

	listing.applySort()

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

	page := &page.Page{
		Info: &page.Info{
			Name:   listing.Name,
			Path:   i.VirtualPath,
			IsDir:  true,
			User:   u,
			Config: c,
			Data:   listing,
		},
	}

	if r.Header.Get("Minimal") == "true" {
		page.Minimal = true
	}

	return page.PrintAsHTML(w, "listing")
}

func (i Info) loadDirectoryContents(file http.File, path string, u *config.User) (*Listing, error) {
	files, err := file.Readdir(-1)
	if err != nil {
		return nil, err
	}

	listing := directoryListing(files, i.VirtualPath, path, u)
	return &listing, nil
}

func directoryListing(files []os.FileInfo, urlPath string, basePath string, u *config.User) Listing {
	var (
		fileinfos           []Info
		dirCount, fileCount int
	)

	for _, f := range files {
		name := f.Name()

		if f.IsDir() {
			name += "/"
			dirCount++
		} else {
			fileCount++
		}

		// Absolute URL
		url := url.URL{Path: basePath + name}
		fileinfos = append(fileinfos, Info{
			FileInfo:    f,
			URL:         url.String(),
			UserAllowed: u.Allowed(url.String()),
		})
	}

	return Listing{
		Name:     path.Base(urlPath),
		Path:     urlPath,
		Items:    fileinfos,
		NumDirs:  dirCount,
		NumFiles: fileCount,
	}
}

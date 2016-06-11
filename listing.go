package filemanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/mholt/caddy/caddyhttp/staticfiles"
)

// A Listing is the context used to fill out a template.
type Listing struct {
	// The name of the directory (the last element of the path)
	Name string

	// The full path of the request
	Path string

	// Whether the parent directory is browsable
	CanGoUp bool

	// The items (files and folders) in the path
	Items []FileInfo

	// The number of directories in the listing
	NumDirs int

	// The number of files (items that aren't directories) in the listing
	NumFiles int

	// Which sorting order is used
	Sort string

	// And which order
	Order string

	// If ≠0 then Items have been limited to that many elements
	ItemsLimitedTo int

	// Optional custom variables for use in browse templates
	User interface{}

	httpserver.Context
}

func (f FileManager) loadDirectoryContents(requestedFilepath http.File, urlPath string) (*Listing, bool, error) {
	files, err := requestedFilepath.Readdir(-1)
	if err != nil {
		return nil, false, err
	}

	// Determine if user can browse up another folder
	// TODO: review this
	var canGoUp bool
	curPathDir := path.Dir(strings.TrimSuffix(urlPath, "/"))
	for _, other := range f.Configs {
		if strings.HasPrefix(curPathDir, other.PathScope) {
			canGoUp = true
			break
		}
	}

	// Assemble listing of directory contents
	listing, hasIndex := directoryListing(files, canGoUp, urlPath)
	return &listing, hasIndex, nil
}

// ServeListing returns a formatted view of 'requestedFilepath' contents'.
func (f FileManager) ServeListing(w http.ResponseWriter, r *http.Request, requestedFilepath http.File, bc *Config) (int, error) {
	listing, containsIndex, err := f.loadDirectoryContents(requestedFilepath, r.URL.Path)
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

	if containsIndex && !f.IgnoreIndexes { // directory isn't browsable
		return f.Next.ServeHTTP(w, r)
	}
	listing.Context = httpserver.Context{
		Root: bc.Root,
		Req:  r,
		URL:  r.URL,
	}
	listing.User = bc.Variables

	// Copy the query values into the Listing struct
	var limit int
	listing.Sort, listing.Order, limit, err = f.handleSortOrder(w, r, bc.PathScope)
	if err != nil {
		return http.StatusBadRequest, err
	}

	listing.applySort()

	if limit > 0 && limit <= len(listing.Items) {
		listing.Items = listing.Items[:limit]
		listing.ItemsLimitedTo = limit
	}

	var buf *bytes.Buffer
	acceptHeader := strings.ToLower(strings.Join(r.Header["Accept"], ","))

	if strings.Contains(acceptHeader, "application/json") {
		if buf, err = f.formatAsJSON(listing, bc); err != nil {
			return http.StatusInternalServerError, err
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	} else {
		page := &Page{
			Name:   listing.Name,
			Path:   listing.Path,
			Config: bc,
			Data:   listing,
		}

		if buf, err = f.formatAsHTML(page, "listing"); err != nil {
			return http.StatusInternalServerError, err
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	buf.WriteTo(w)
	return http.StatusOK, nil
}

func (f FileManager) formatAsJSON(listing *Listing, bc *Config) (*bytes.Buffer, error) {
	marsh, err := json.Marshal(listing.Items)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.Write(marsh)
	return buf, err
}

func directoryListing(files []os.FileInfo, canGoUp bool, urlPath string) (Listing, bool) {
	var (
		fileinfos           []FileInfo
		dirCount, fileCount int
		hasIndexFile        bool
	)

	for _, f := range files {
		name := f.Name()

		for _, indexName := range staticfiles.IndexPages {
			if name == indexName {
				hasIndexFile = true
				break
			}
		}

		if f.IsDir() {
			name += "/"
			dirCount++
		} else {
			fileCount++
		}

		url := url.URL{Path: "./" + name} // prepend with "./" to fix paths with ':' in the name

		fileinfos = append(fileinfos, FileInfo{
			IsDir:   f.IsDir(),
			Name:    f.Name(),
			Size:    f.Size(),
			URL:     url.String(),
			ModTime: f.ModTime().UTC(),
			Mode:    f.Mode(),
		})
	}

	return Listing{
		Name:     path.Base(urlPath),
		Path:     urlPath,
		CanGoUp:  canGoUp,
		Items:    fileinfos,
		NumDirs:  dirCount,
		NumFiles: fileCount,
	}, hasIndexFile
}

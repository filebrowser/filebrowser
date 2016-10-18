package filemanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
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
	ItemsLimitedTo     int
	httpserver.Context `json:"-"`
}

// handleSortOrder gets and stores for a Listing the 'sort' and 'order',
// and reads 'limit' if given. The latter is 0 if not given. Sets cookies.
func handleSortOrder(w http.ResponseWriter, r *http.Request, scope string) (sort string, order string, limit int, err error) {
	sort, order, limitQuery := r.URL.Query().Get("sort"), r.URL.Query().Get("order"), r.URL.Query().Get("limit")

	// If the query 'sort' or 'order' is empty, use defaults or any values previously saved in Cookies
	switch sort {
	case "":
		sort = "name"
		if sortCookie, sortErr := r.Cookie("sort"); sortErr == nil {
			sort = sortCookie.Value
		}
	case "name", "size", "type":
		http.SetCookie(w, &http.Cookie{Name: "sort", Value: sort, Path: scope, Secure: r.TLS != nil})
	}

	switch order {
	case "":
		order = "asc"
		if orderCookie, orderErr := r.Cookie("order"); orderErr == nil {
			order = orderCookie.Value
		}
	case "asc", "desc":
		http.SetCookie(w, &http.Cookie{Name: "order", Value: order, Path: scope, Secure: r.TLS != nil})
	}

	if limitQuery != "" {
		limit, err = strconv.Atoi(limitQuery)
		if err != nil { // if the 'limit' query can't be interpreted as a number, return err
			return
		}
	}

	return
}

// Implement sorting for Listing
type byName Listing
type bySize Listing
type byTime Listing

// By Name
func (l byName) Len() int      { return len(l.Items) }
func (l byName) Swap(i, j int) { l.Items[i], l.Items[j] = l.Items[j], l.Items[i] }

// Treat upper and lower case equally
func (l byName) Less(i, j int) bool {
	if l.Items[i].IsDir() && !l.Items[j].IsDir() {
		return true
	}

	if !l.Items[i].IsDir() && l.Items[j].IsDir() {
		return false
	}

	return strings.ToLower(l.Items[i].Name()) < strings.ToLower(l.Items[j].Name())
}

// By Size
func (l bySize) Len() int      { return len(l.Items) }
func (l bySize) Swap(i, j int) { l.Items[i], l.Items[j] = l.Items[j], l.Items[i] }

const directoryOffset = -1 << 31 // = math.MinInt32
func (l bySize) Less(i, j int) bool {
	iSize, jSize := l.Items[i].Size(), l.Items[j].Size()
	if l.Items[i].IsDir() {
		iSize = directoryOffset + iSize
	}
	if l.Items[j].IsDir() {
		jSize = directoryOffset + jSize
	}
	return iSize < jSize
}

// By Time
func (l byTime) Len() int           { return len(l.Items) }
func (l byTime) Swap(i, j int)      { l.Items[i], l.Items[j] = l.Items[j], l.Items[i] }
func (l byTime) Less(i, j int) bool { return l.Items[i].ModTime().Before(l.Items[j].ModTime()) }

// Add sorting method to "Listing"
// it will apply what's in ".Sort" and ".Order"
func (l Listing) applySort() {
	// Check '.Order' to know how to sort
	if l.Order == "desc" {
		switch l.Sort {
		case "name":
			sort.Sort(sort.Reverse(byName(l)))
		case "size":
			sort.Sort(sort.Reverse(bySize(l)))
		case "time":
			sort.Sort(sort.Reverse(byTime(l)))
		default:
			// If not one of the above, do nothing
			return
		}
	} else { // If we had more Orderings we could add them here
		switch l.Sort {
		case "name":
			sort.Sort(byName(l))
		case "size":
			sort.Sort(bySize(l))
		case "time":
			sort.Sort(byTime(l))
		default:
			sort.Sort(byName(l))
			return
		}
	}
}

func (i *FileInfo) serveListing(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.User) (int, error) {
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

func (i FileInfo) loadDirectoryContents(file http.File, path string, u *config.User) (*Listing, error) {
	files, err := file.Readdir(-1)
	if err != nil {
		return nil, err
	}

	listing := directoryListing(files, i.VirtualPath, path, u)
	return &listing, nil
}

func directoryListing(files []os.FileInfo, urlPath string, basePath string, u *config.User) Listing {
	var (
		fileinfos           []FileInfo
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
		fileinfos = append(fileinfos, FileInfo{
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

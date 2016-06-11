package filemanager

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

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
	ItemsLimitedTo int
	httpserver.Context
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
	return strings.ToLower(l.Items[i].Name) < strings.ToLower(l.Items[j].Name)
}

// By Size
func (l bySize) Len() int      { return len(l.Items) }
func (l bySize) Swap(i, j int) { l.Items[i], l.Items[j] = l.Items[j], l.Items[i] }

const directoryOffset = -1 << 31 // = math.MinInt32
func (l bySize) Less(i, j int) bool {
	iSize, jSize := l.Items[i].Size, l.Items[j].Size
	if l.Items[i].IsDir {
		iSize = directoryOffset + iSize
	}
	if l.Items[j].IsDir {
		jSize = directoryOffset + jSize
	}
	return iSize < jSize
}

// By Time
func (l byTime) Len() int           { return len(l.Items) }
func (l byTime) Swap(i, j int)      { l.Items[i], l.Items[j] = l.Items[j], l.Items[i] }
func (l byTime) Less(i, j int) bool { return l.Items[i].ModTime.Before(l.Items[j].ModTime) }

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
			// If not one of the above, do nothing
			return
		}
	}
}

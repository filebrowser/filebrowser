package file

import (
	"context"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/hacdias/caddy-filemanager/config"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// A Listing is the context used to fill out a template.
type Listing struct {
	// The name of the directory (the last element of the path)
	Name string
	// The full path of the request relatively to a File System
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

// GetListing gets the information about a specific directory and its files.
func GetListing(u *config.User, filePath string, baseURL string) (*Listing, error) {
	// Gets the directory information using the Virtual File System of
	// the user configuration.
	file, err := u.FileSystem.OpenFile(context.TODO(), filePath, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Reads the directory and gets the information about the files.
	files, err := file.Readdir(-1)
	if err != nil {
		return nil, err
	}

	var (
		fileinfos           []Info
		dirCount, fileCount int
	)

	for _, f := range files {
		name := f.Name()
		allowed := u.Allowed("/" + name)

		if !allowed {
			continue
		}

		if f.IsDir() {
			name += "/"
			dirCount++
		} else {
			fileCount++
		}

		// Absolute URL
		url := url.URL{Path: baseURL + name}

		i := Info{
			Name:        f.Name(),
			Size:        f.Size(),
			ModTime:     f.ModTime(),
			Mode:        f.Mode(),
			IsDir:       f.IsDir(),
			URL:         url.String(),
			UserAllowed: allowed,
		}
		i.RetrieveFileType()

		fileinfos = append(fileinfos, i)
	}

	return &Listing{
		Name:     path.Base(filePath),
		Path:     filePath,
		Items:    fileinfos,
		NumDirs:  dirCount,
		NumFiles: fileCount,
	}, nil
}

// ApplySort applies the sort order using .Order and .Sort
func (l Listing) ApplySort() {
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

// Implement sorting for Listing
type byName Listing
type bySize Listing
type byTime Listing

// By Name
func (l byName) Len() int {
	return len(l.Items)
}

func (l byName) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}

// Treat upper and lower case equally
func (l byName) Less(i, j int) bool {
	if l.Items[i].IsDir && !l.Items[j].IsDir {
		return true
	}

	if !l.Items[i].IsDir && l.Items[j].IsDir {
		return false
	}

	return strings.ToLower(l.Items[i].Name) < strings.ToLower(l.Items[j].Name)
}

// By Size
func (l bySize) Len() int {
	return len(l.Items)
}

func (l bySize) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}

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
func (l byTime) Len() int {
	return len(l.Items)
}
func (l byTime) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}
func (l byTime) Less(i, j int) bool {
	return l.Items[i].ModTime.Before(l.Items[j].ModTime)
}

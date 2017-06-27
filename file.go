package filemanager

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
)

var (
	errInvalidOption = errors.New("Invalid option")
)

// fileInfo contains the information about a particular file or directory.
type fileInfo struct {
	// Used to store the file's content temporarily.
	content []byte
	// The name of the file.
	Name string
	// The Size of the file.
	Size int64
	// The absolute URL.
	URL string
	// The extension of the file.
	Extension string
	// The last modified time.
	ModTime time.Time
	// The File Mode.
	Mode os.FileMode
	// Indicates if this file is a directory.
	IsDir bool
	// Absolute path.
	Path string
	// Relative path to user's virtual File System.
	VirtualPath string
	// Indicates the file content type: video, text, image, music or blob.
	Type string
}

// A listing is the context used to fill out a template.
type listing struct {
	// The name of the directory (the last element of the path).
	Name string
	// The full path of the request relatively to a File System.
	Path string
	// The items (files and folders) in the path.
	Items []fileInfo
	// The number of directories in the listing.
	NumDirs int
	// The number of files (items that aren't directories) in the listing.
	NumFiles int
	// Which sorting order is used.
	Sort string
	// And which order.
	Order string
	// If ≠0 then Items have been limited to that many elements.
	ItemsLimitedTo int
	Display        string
	// Indicates if we're showing a preview in this Listing
	Preview bool
	// File to preview if Preview is true
	PreviewItem *fileInfo
}

// getInfo gets the file information and, in case of error, returns the
// respective HTTP error code
func getInfo(url *url.URL, c *FileManager, u *User) (*fileInfo, error) {
	var err error

	i := &fileInfo{URL: c.RootURL() + url.Path}
	i.VirtualPath = url.Path
	i.VirtualPath = strings.TrimPrefix(i.VirtualPath, "/")
	i.VirtualPath = "/" + i.VirtualPath

	i.Path = u.scope + i.VirtualPath
	i.Path = filepath.Clean(i.Path)

	info, err := os.Stat(i.Path)
	if err != nil {
		return i, err
	}

	i.Name = info.Name()
	i.ModTime = info.ModTime()
	i.Mode = info.Mode()
	i.IsDir = info.IsDir()
	i.Size = info.Size()
	i.Extension = filepath.Ext(i.Name)
	return i, nil
}

// getListing gets the information about a specific directory and its files.
func getListing(u *User, filePath string, baseURL string) (*listing, error) {
	// Gets the directory information using the Virtual File System of
	// the user configuration.
	file, err := u.fileSystem.OpenFile(context.TODO(), filePath, os.O_RDONLY, 0)
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
		fileinfos           []fileInfo
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

		i := fileInfo{
			Name:    f.Name(),
			Size:    f.Size(),
			ModTime: f.ModTime(),
			Mode:    f.Mode(),
			IsDir:   f.IsDir(),
			URL:     url.String(),
		}
		i.RetrieveFileType()

		fileinfos = append(fileinfos, i)
	}

	return &listing{
		Name:     path.Base(filePath),
		Path:     filePath,
		Items:    fileinfos,
		NumDirs:  dirCount,
		NumFiles: fileCount,
	}, nil
}

// RetrieveFileType obtains the mimetype and converts it to a simple
// type nomenclature.
func (i *fileInfo) RetrieveFileType() error {
	// Tries to get the file mimetype using its extension.
	mimetype := mime.TypeByExtension(i.Extension)

	if mimetype == "" {
		err := i.Read()
		if err != nil {
			return err
		}

		// Tries to get the file mimetype using its first
		// 512 bytes.
		mimetype = http.DetectContentType(i.content)
	}

	if strings.HasPrefix(mimetype, "video") {
		i.Type = "video"
		return nil
	}

	if strings.HasPrefix(mimetype, "audio") {
		i.Type = "audio"
		return nil
	}

	if strings.HasPrefix(mimetype, "image") {
		i.Type = "image"
		return nil
	}

	if strings.HasPrefix(mimetype, "text") {
		i.Type = "text"
		return nil
	}

	if strings.HasPrefix(mimetype, "application/javascript") {
		i.Type = "text"
		return nil
	}

	// If the type isn't text (and is blob for example), it will check some
	// common types that are mistaken not to be text.
	for _, extension := range textExtensions {
		if strings.HasSuffix(i.Name, extension) {
			i.Type = "text"
			return nil
		}
	}

	i.Type = "blob"
	return nil
}

// Reads the file.
func (i *fileInfo) Read() error {
	if len(i.content) != 0 {
		return nil
	}

	var err error
	i.content, err = ioutil.ReadFile(i.Path)
	if err != nil {
		return err
	}
	return nil
}

func (i fileInfo) Checksum(kind string) (string, error) {
	file, err := os.Open(i.Path)
	if err != nil {
		return "", err
	}

	defer file.Close()

	var h hash.Hash

	switch kind {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return "", errInvalidOption
	}

	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// StringifyContent returns a string with the file content.
func (i fileInfo) StringifyContent() string {
	return string(i.content)
}

// HumanSize returns the size of the file as a human-readable string
// in IEC format (i.e. power of 2 or base 1024).
func (i fileInfo) HumanSize() string {
	return humanize.IBytes(uint64(i.Size))
}

// HumanModTime returns the modified time of the file as a human-readable string.
func (i fileInfo) HumanModTime(format string) string {
	return i.ModTime.Format(format)
}

// CanBeEdited checks if the extension of a file is supported by the editor
func (i fileInfo) CanBeEdited() bool {
	return i.Type == "text"
}

// ApplySort applies the sort order using .Order and .Sort
func (l listing) ApplySort() {
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

// Implement sorting for listing
type byName listing
type bySize listing
type byTime listing

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

var textExtensions = [...]string{
	".md", ".markdown", ".mdown", ".mmark",
	".asciidoc", ".adoc", ".ad",
	".rst",
	".json", ".toml", ".yaml", ".csv", ".xml", ".rss", ".conf", ".ini",
	".tex", ".sty",
	".css", ".sass", ".scss",
	".js",
	".html",
	".txt", ".rtf",
	".sh", ".bash", ".ps1", ".bat", ".cmd",
	".php", ".pl", ".py",
	"Caddyfile",
	".c", ".cc", ".h", ".hh", ".cpp", ".hpp", ".f90",
	".f", ".bas", ".d", ".ada", ".nim", ".cr", ".java", ".cs", ".vala", ".vapi",
}

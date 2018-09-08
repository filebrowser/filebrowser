package lib

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gohugoio/hugo/parser"
	"github.com/maruel/natural"
)

// File contains the information about a particular file or directory.
type File struct {
	// Indicates the Kind of view on the front-end (Listing, editor or preview).
	Kind string `json:"kind"`
	// The name of the file.
	Name string `json:"name"`
	// The Size of the file.
	Size int64 `json:"size"`
	// The absolute URL.
	URL string `json:"url"`
	// The extension of the file.
	Extension string `json:"extension"`
	// The last modified time.
	ModTime time.Time `json:"modified"`
	// The File Mode.
	Mode os.FileMode `json:"mode"`
	// Indicates if this file is a directory.
	IsDir bool `json:"isDir"`
	// Absolute path.
	Path string `json:"path"`
	// Relative path to user's virtual File System.
	VirtualPath string `json:"virtualPath"`
	// Indicates the file content type: video, text, image, music or blob.
	Type string `json:"type"`
	// Stores the content of a text file.
	Content string `json:"content,omitempty"`

	*Listing `json:",omitempty"`

	Metadata string `json:"metadata,omitempty"`
	Language string `json:"language,omitempty"`
}

// A Listing is the context used to fill out a template.
type Listing struct {
	// The items (files and folders) in the path.
	Items []*File `json:"items"`
	// The number of directories in the Listing.
	NumDirs int `json:"numDirs"`
	// The number of files (items that aren't directories) in the Listing.
	NumFiles int `json:"numFiles"`
	// Which sorting order is used.
	Sort string `json:"sort"`
	// And which order.
	Order string `json:"order"`
}

// GetInfo gets the file information and, in case of error, returns the
// respective HTTP error code
func GetInfo(url *url.URL, c *FileBrowser, u *User) (*File, error) {
	var err error

	i := &File{
		URL:         "/files" + url.String(),
		VirtualPath: url.Path,
		Path:        filepath.Join(u.Scope, url.Path),
	}

	info, err := u.FileSystem.Stat(url.Path)
	if err != nil {
		return i, err
	}

	i.Name = info.Name()
	i.ModTime = info.ModTime()
	i.Mode = info.Mode()
	i.IsDir = info.IsDir()
	i.Size = info.Size()
	i.Extension = filepath.Ext(i.Name)

	if i.IsDir && !strings.HasSuffix(i.URL, "/") {
		i.URL += "/"
	}

	return i, nil
}

// GetListing gets the information about a specific directory and its files.
func (i *File) GetListing(u *User, r *http.Request) error {
	// Gets the directory information using the Virtual File System of
	// the user configuration.
	f, err := u.FileSystem.OpenFile(i.VirtualPath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	// Reads the directory and gets the information about the files.
	files, err := f.Readdir(-1)
	if err != nil {
		return err
	}

	var (
		fileinfos           []*File
		dirCount, fileCount int
	)

	baseurl, err := url.PathUnescape(i.URL)
	if err != nil {
		return err
	}

	for _, f := range files {
		name := f.Name()
		allowed := u.Allowed("/" + name)

		if !allowed {
			continue
		}

		if strings.HasPrefix(f.Mode().String(), "L") {
			// It's a symbolic link. We try to follow it. If it doesn't work,
			// we stay with the link information instead if the target's.
			info, err := os.Stat(f.Name())
			if err == nil {
				f = info
			}
		}

		if f.IsDir() {
			name += "/"
			dirCount++
		} else {
			fileCount++
		}

		// Absolute URL
		url := url.URL{Path: baseurl + name}

		i := &File{
			Name:        f.Name(),
			Size:        f.Size(),
			ModTime:     f.ModTime(),
			Mode:        f.Mode(),
			IsDir:       f.IsDir(),
			URL:         url.String(),
			Extension:   filepath.Ext(name),
			VirtualPath: filepath.Join(i.VirtualPath, name),
			Path:        filepath.Join(i.Path, name),
		}

		i.GetFileType(false)
		fileinfos = append(fileinfos, i)
	}

	i.Listing = &Listing{
		Items:    fileinfos,
		NumDirs:  dirCount,
		NumFiles: fileCount,
	}

	return nil
}

// GetEditor gets the editor based on a Info struct
func (i *File) GetEditor() error {
	i.Language = editorLanguage(i.Extension)
	// If the editor will hold only content, leave now.
	if editorMode(i.Language) == "content" {
		return nil
	}

	// If the file doesn't have any kind of metadata, leave now.
	if !hasRune(i.Content) {
		return nil
	}

	buffer := bytes.NewBuffer([]byte(i.Content))
	page, err := parser.ReadFrom(buffer)

	// If there is an error, just ignore it and return nil.
	// This way, the file can be served for editing.
	if err != nil {
		return nil
	}

	i.Content = strings.TrimSpace(string(page.Content()))
	i.Metadata = strings.TrimSpace(string(page.FrontMatter()))
	return nil
}

// GetFileType obtains the mimetype and converts it to a simple
// type nomenclature.
func (i *File) GetFileType(checkContent bool) error {
	var content []byte
	var err error

	// Tries to get the file mimetype using its extension.
	mimetype := mime.TypeByExtension(i.Extension)

	if mimetype == "" && checkContent {
		file, err := os.Open(i.Path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Only the first 512 bytes are used to sniff the content type.
		buffer := make([]byte, 512)
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		// Tries to get the file mimetype using its first
		// 512 bytes.
		mimetype = http.DetectContentType(buffer[:n])
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
		goto End
	}

	if strings.HasPrefix(mimetype, "application/javascript") {
		i.Type = "text"
		goto End
	}

	// If the type isn't text (and is blob for example), it will check some
	// common types that are mistaken not to be text.
	if isInTextExtensions(i.Name) {
		i.Type = "text"
	} else {
		i.Type = "blob"
	}

End:
	// If the file type is text, save its content.
	if i.Type == "text" {
		if len(content) == 0 {
			content, err = ioutil.ReadFile(i.Path)
			if err != nil {
				return err
			}
		}

		i.Content = string(content)
	}

	return nil
}

// Checksum retrieves the checksum of a file.
func (i File) Checksum(algo string) (string, error) {
	file, err := os.Open(i.Path)
	if err != nil {
		return "", err
	}

	defer file.Close()

	var h hash.Hash

	switch algo {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return "", ErrInvalidOption
	}

	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// CanBeEdited checks if the extension of a file is supported by the editor
func (i File) CanBeEdited() bool {
	return i.Type == "text"
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
		case "modified":
			sort.Sort(sort.Reverse(byModified(l)))
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
		case "modified":
			sort.Sort(byModified(l))
		default:
			sort.Sort(byName(l))
			return
		}
	}
}

// Implement sorting for Listing
type byName Listing
type bySize Listing
type byModified Listing

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

	return natural.Less(l.Items[i].Name, l.Items[j].Name)
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

// By Modified
func (l byModified) Len() int {
	return len(l.Items)
}

func (l byModified) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}

func (l byModified) Less(i, j int) bool {
	iModified, jModified := l.Items[i].ModTime, l.Items[j].ModTime
	return iModified.Sub(jModified) < 0
}

// textExtensions is the sorted list of text extensions which
// can be edited.
var textExtensions = []string{
	".ad", ".ada", ".adoc", ".asciidoc",
	".bas", ".bash", ".bat",
	".c", ".cc", ".cmd", ".conf", ".cpp", ".cr", ".cs", ".css", ".csv",
	".d",
	".f", ".f90",
	".h", ".hh", ".hpp", ".htaccess", ".html",
	".ini",
	".java", ".js", ".json",
	".markdown", ".md", ".mdown", ".mmark",
	".nim",
	".php", ".pl", ".ps1", ".py",
	".rss", ".rst", ".rtf",
	".sass", ".scss", ".sh", ".sty",
	".tex", ".tml", ".toml", ".txt",
	".vala", ".vapi",
	".xml",
	".yaml", ".yml",
	"Caddyfile",
}

// isInTextExtensions checks if a file can be edited by its extensions.
func isInTextExtensions(name string) bool {
	search := filepath.Ext(name)
	if search == "" {
		search = name
	}

	i := sort.SearchStrings(textExtensions, search)
	return i < len(textExtensions) && textExtensions[i] == search
}

// hasRune checks if the file has the frontmatter rune
func hasRune(file string) bool {
	return strings.HasPrefix(file, "---") ||
		strings.HasPrefix(file, "+++") ||
		strings.HasPrefix(file, "{")
}

func editorMode(language string) string {
	switch language {
	case "markdown", "asciidoc", "rst":
		return "content+metadata"
	}

	return "content"
}

func editorLanguage(mode string) string {
	mode = strings.TrimPrefix(mode, ".")

	switch mode {
	case "md", "markdown", "mdown", "mmark":
		mode = "markdown"
	case "yml":
		mode = "yaml"
	case "asciidoc", "adoc", "ad":
		mode = "asciidoc"
	case "rst":
		mode = "rst"
	case "html", "htm", "xml":
		mode = "htmlmixed"
	case "js":
		mode = "javascript"
	case "go":
		mode = "golang"
	case "":
		mode = "text"
	}

	return mode
}

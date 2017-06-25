package filemanager

import (
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
)

// FileInfo contains the information about a particular file or directory
type FileInfo struct {
	Name        string
	Size        int64
	URL         string
	Extension   string
	ModTime     time.Time
	Mode        os.FileMode
	IsDir       bool
	Path        string // Relative path to Current Working Directory
	VirtualPath string // Relative path to user's virtual File System
	Mimetype    string
	Type        string
	UserAllowed bool // Indicates if the user has enough permissions

	content []byte
}

// GetInfo gets the file information and, in case of error, returns the
// respective HTTP error code
func GetInfo(url *url.URL, c *FileManager, u *User) (*FileInfo, error) {
	var err error

	i := &FileInfo{URL: c.PrefixURL + url.Path}
	i.VirtualPath = strings.Replace(url.Path, c.BaseURL, "", 1)
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

// RetrieveFileType obtains the mimetype and a simplified internal Type
// using the first 512 bytes from the file.
func (i *FileInfo) RetrieveFileType() error {
	i.Mimetype = mime.TypeByExtension(i.Extension)

	if i.Mimetype == "" {
		err := i.Read()
		if err != nil {
			return err
		}

		i.Mimetype = http.DetectContentType(i.content)
	}

	if strings.HasPrefix(i.Mimetype, "video") {
		i.Type = "video"
		return nil
	}

	if strings.HasPrefix(i.Mimetype, "audio") {
		i.Type = "audio"
		return nil
	}

	if strings.HasPrefix(i.Mimetype, "image") {
		i.Type = "image"
		return nil
	}

	if strings.HasPrefix(i.Mimetype, "text") {
		i.Type = "text"
		return nil
	}

	if strings.HasPrefix(i.Mimetype, "application/javascript") {
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
func (i *FileInfo) Read() error {
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

// StringifyContent returns the string version of Raw
func (i FileInfo) StringifyContent() string {
	return string(i.content)
}

// HumanSize returns the size of the file as a human-readable string
// in IEC format (i.e. power of 2 or base 1024).
func (i FileInfo) HumanSize() string {
	return humanize.IBytes(uint64(i.Size))
}

// HumanModTime returns the modified time of the file as a human-readable string.
func (i FileInfo) HumanModTime(format string) string {
	return i.ModTime.Format(format)
}

// CanBeEdited checks if the extension of a file is supported by the editor
func (i FileInfo) CanBeEdited() bool {
	return i.Type == "text"
}

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

// fileInfo contains the information about a particular file or directory.
type fileInfo struct {
	// Used to store the file's content temporarily.
	content []byte

	Name      string
	Size      int64
	URL       string
	Extension string
	ModTime   time.Time
	Mode      os.FileMode
	IsDir     bool

	// Absolute path.
	Path string

	// Relative path to user's virtual File System.
	VirtualPath string

	// Indicates the file content type: video, text, image, music or blob.
	Type string
}

// getInfo gets the file information and, in case of error, returns the
// respective HTTP error code
func getInfo(url *url.URL, c *FileManager, u *User) (*fileInfo, error) {
	var err error

	i := &fileInfo{URL: c.PrefixURL + url.Path}
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

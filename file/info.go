package file

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/utils"
)

// Info contains the information about a particular file or directory
type Info struct {
	os.FileInfo
	URL         string
	Path        string // Relative path to Caddyfile
	VirtualPath string // Relative path to u.FileSystem
	Mimetype    string
	Content     []byte
	Type        string
	UserAllowed bool // Indicates if the user has enough permissions
}

// GetInfo gets the file information and, in case of error, returns the
// respective HTTP error code
func GetInfo(url *url.URL, c *config.Config, u *config.User) (*Info, int, error) {
	var err error

	i := &Info{URL: url.Path}
	i.VirtualPath = strings.Replace(url.Path, c.BaseURL, "", 1)
	i.VirtualPath = strings.TrimPrefix(i.VirtualPath, "/")
	i.VirtualPath = "/" + i.VirtualPath

	i.Path = u.Scope + i.VirtualPath
	i.Path = strings.Replace(i.Path, "\\", "/", -1)
	i.Path = filepath.Clean(i.Path)

	i.FileInfo, err = os.Stat(i.Path)
	if err != nil {
		return i, utils.ErrorToHTTPCode(err, false), err
	}

	return i, 0, nil
}

func (i *Info) Read() error {
	var err error
	i.Content, err = ioutil.ReadFile(i.Path)
	if err != nil {
		return err
	}
	i.Mimetype = http.DetectContentType(i.Content)
	i.Type = simplifyMediaType(i.Mimetype)
	return nil
}

// StringifyContent returns the string version of Raw
func (i Info) StringifyContent() string {
	return string(i.Content)
}

// HumanSize returns the size of the file as a human-readable string
// in IEC format (i.e. power of 2 or base 1024).
func (i Info) HumanSize() string {
	return humanize.IBytes(uint64(i.Size()))
}

// HumanModTime returns the modified time of the file as a human-readable string.
func (i Info) HumanModTime(format string) string {
	return i.ModTime().Format(format)
}

func simplifyMediaType(name string) string {
	if strings.HasPrefix(name, "video") {
		return "video"
	}

	if strings.HasPrefix(name, "audio") {
		return "audio"
	}

	if strings.HasPrefix(name, "image") {
		return "image"
	}

	if strings.HasPrefix(name, "text") {
		return "text"
	}

	if strings.HasPrefix(name, "application/javascript") {
		return "text"
	}

	return "blob"
}

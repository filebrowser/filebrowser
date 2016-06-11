package filemanager

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

// FileInfo is the information about a particular file or directory
type FileInfo struct {
	IsDir   bool
	Name    string
	Size    int64
	URL     string
	ModTime time.Time
	Mode    os.FileMode
	Path    string
}

// GetFileInfo gets the file information and, in case of error, returns the
// respective HTTP error code
func GetFileInfo(url *url.URL, c *Config) (*FileInfo, int, error) {
	var err error

	path := strings.Replace(url.Path, c.BaseURL, c.PathScope, 1)
	path = filepath.Clean(path)
	path = strings.Replace(path, "\\", "/", -1)

	file := &FileInfo{Path: path}
	f, err := c.Root.Open("/" + path)
	if err != nil {
		return file, ErrorToHTTPCode(err), err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return file, ErrorToHTTPCode(err), err
	}

	file.IsDir = info.IsDir()
	file.ModTime = info.ModTime()
	file.Name = info.Name()
	file.Size = info.Size()
	file.URL = url.Path

	return file, 0, nil
}

// GetExtendedFileInfo is used to get extra parameters for FileInfo struct
func (fi FileInfo) GetExtendedFileInfo() error {
	// TODO: do this!
	return nil
}

// HumanSize returns the size of the file as a human-readable string
// in IEC format (i.e. power of 2 or base 1024).
func (fi FileInfo) HumanSize() string {
	return humanize.IBytes(uint64(fi.Size))
}

// HumanModTime returns the modified time of the file as a human-readable string.
func (fi FileInfo) HumanModTime(format string) string {
	return fi.ModTime.Format(format)
}

// Delete handles the delete requests
func (fi FileInfo) Delete() (int, error) {
	var err error

	// If it's a directory remove all the contents inside
	if fi.IsDir {
		err = os.RemoveAll(fi.Path)
	} else {
		err = os.Remove(fi.Path)
	}

	if err != nil {
		return ErrorToHTTPCode(err), err
	}

	return http.StatusOK, nil
}

// Rename function is used tor rename a file or a directory
func (fi FileInfo) Rename(w http.ResponseWriter, r *http.Request) (int, error) {
	newname := r.Header.Get("Rename-To")
	if newname == "" {
		return http.StatusBadRequest, nil
	}

	newpath := filepath.Clean(newname)
	newpath = strings.Replace(fi.Path, fi.Name, newname, 1)

	if err := os.Rename(fi.Path, newpath); err != nil {
		return ErrorToHTTPCode(err), err
	}

	http.Redirect(w, r, strings.Replace(fi.URL, fi.Name, newname, 1), http.StatusTemporaryRedirect)
	return 0, nil
}

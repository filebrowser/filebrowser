package directory

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hacdias/caddy-filemanager/config"
	p "github.com/hacdias/caddy-filemanager/page"
	"github.com/hacdias/caddy-filemanager/utils/errors"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// Info is the information about a particular file or directory
type Info struct {
	IsDir    bool
	Name     string
	Size     int64
	URL      string
	Path     string // The relative Path of the file/directory relative to Caddyfile.
	RootPath string // The Path of the file/directory on http.FileSystem.
	ModTime  time.Time
	Mode     os.FileMode
	Mimetype string
	Content  string
	Raw      []byte
	Type     string
}

// GetInfo gets the file information and, in case of error, returns the
// respective HTTP error code
func GetInfo(url *url.URL, c *config.Config) (*Info, int, error) {
	var err error

	rootPath := strings.Replace(url.Path, c.BaseURL, "", 1)
	rootPath = strings.TrimPrefix(rootPath, "/")
	rootPath = "/" + rootPath

	relpath := c.PathScope + rootPath
	relpath = strings.Replace(relpath, "\\", "/", -1)
	relpath = filepath.Clean(relpath)

	file := &Info{
		URL:      url.Path,
		RootPath: rootPath,
		Path:     relpath,
	}
	f, err := c.Root.Open(rootPath)
	if err != nil {
		return file, errors.ToHTTPCode(err), err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return file, errors.ToHTTPCode(err), err
	}

	file.IsDir = info.IsDir()
	file.ModTime = info.ModTime()
	file.Name = info.Name()
	file.Size = info.Size()
	return file, 0, nil
}

// GetExtendedInfo is used to get extra parameters for FileInfo struct
func (i *Info) GetExtendedInfo() error {
	err := i.Read()
	if err != nil {
		return err
	}

	i.Type = SimplifyMimeType(i.Mimetype)
	return nil
}

// Read is used to read a file and store its content
func (i *Info) Read() error {
	raw, err := ioutil.ReadFile(i.Path)
	if err != nil {
		return err
	}
	i.Mimetype = http.DetectContentType(raw)
	i.Content = string(raw)
	i.Raw = raw
	return nil
}

// HumanSize returns the size of the file as a human-readable string
// in IEC format (i.e. power of 2 or base 1024).
func (i Info) HumanSize() string {
	return humanize.IBytes(uint64(i.Size))
}

// HumanModTime returns the modified time of the file as a human-readable string.
func (i Info) HumanModTime(format string) string {
	return i.ModTime.Format(format)
}

// Delete handles the delete requests
func (i *Info) Delete() (int, error) {
	var err error

	// If it's a directory remove all the contents inside
	if i.IsDir {
		err = os.RemoveAll(i.Path)
	} else {
		err = os.Remove(i.Path)
	}

	if err != nil {
		return errors.ToHTTPCode(err), err
	}

	return http.StatusOK, nil
}

// Rename function is used tor rename a file or a directory
func (i *Info) Rename(w http.ResponseWriter, r *http.Request) (int, error) {
	newname := r.Header.Get("Rename-To")
	if newname == "" {
		return http.StatusBadRequest, nil
	}

	newpath := filepath.Clean(newname)
	newpath = strings.Replace(i.Path, i.Name, newname, 1)

	if err := os.Rename(i.Path, newpath); err != nil {
		return errors.ToHTTPCode(err), err
	}

	i.Path = newpath
	return http.StatusOK, nil
}

// ServeAsHTML is used to serve single file pages
func (i *Info) ServeAsHTML(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.UserConfig) (int, error) {
	if i.IsDir {
		return i.serveListing(w, r, c, u)
	}

	return i.serveSingleFile(w, r, c, u)
}

func (i *Info) serveSingleFile(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.UserConfig) (int, error) {
	err := i.GetExtendedInfo()
	if err != nil {
		return errors.ToHTTPCode(err), err
	}

	if i.Type == "blob" {
		return i.ServeRawFile(w, r, c)
	}

	page := &p.Page{
		Info: &p.Info{
			Name:   i.Name,
			Path:   i.RootPath,
			IsDir:  false,
			Data:   i,
			User:   u,
			Config: c,
		},
	}

	if CanBeEdited(i.Name) {
		editor, err := i.GetEditor()

		if err != nil {
			return http.StatusInternalServerError, err
		}

		page.Info.Data = editor
		return page.PrintAsHTML(w, "frontmatter", "editor")
	}

	return page.PrintAsHTML(w, "single")
}

func (i *Info) serveListing(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.UserConfig) (int, error) {
	var err error

	file, err := c.Root.Open(i.RootPath)
	if err != nil {
		return errors.ToHTTPCode(err), err
	}
	defer file.Close()

	listing, err := i.loadDirectoryContents(file, c)
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
		Root: c.Root,
		Req:  r,
		URL:  r.URL,
	}

	// Copy the query values into the Listing struct
	var limit int
	listing.Sort, listing.Order, limit, err = handleSortOrder(w, r, c.PathScope)
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

	page := &p.Page{
		Info: &p.Info{
			Name:   listing.Name,
			Path:   i.RootPath,
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

func (i Info) loadDirectoryContents(file http.File, c *config.Config) (*Listing, error) {
	files, err := file.Readdir(-1)
	if err != nil {
		return nil, err
	}

	listing := directoryListing(files, i.RootPath)
	return &listing, nil
}

func directoryListing(files []os.FileInfo, urlPath string) Listing {
	var (
		fileinfos           []Info
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

		url := url.URL{Path: "./" + name} // prepend with "./" to fix paths with ':' in the name

		fileinfos = append(fileinfos, Info{
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
		Items:    fileinfos,
		NumDirs:  dirCount,
		NumFiles: fileCount,
	}
}

// ServeRawFile serves raw files
func (i *Info) ServeRawFile(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	err := i.GetExtendedInfo()
	if err != nil {
		return errors.ToHTTPCode(err), err
	}

	if i.Type != "text" {
		i.Read()
	}

	w.Header().Set("Content-Type", i.Mimetype)
	w.Write([]byte(i.Content))
	return 200, nil
}

// SimplifyMimeType returns the base type of a file
func SimplifyMimeType(name string) string {
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

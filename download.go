package filemanager

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
)

// serveDownload creates an archive in one of the supported formats (zip, tar,
// tar.gz or tar.bz2) and sends it to be downloaded.
func serveDownload(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	query := r.URL.Query().Get("download")

	if !c.fi.IsDir {
		w.Header().Set("Content-Disposition", "attachment; filename="+c.fi.Name)
		http.ServeFile(w, r, c.fi.Path)
		return 0, nil
	}

	files := []string{}
	names := strings.Split(r.URL.Query().Get("files"), ",")

	if len(names) != 0 {
		for _, name := range names {
			name, err := url.QueryUnescape(name)

			if err != nil {
				return http.StatusInternalServerError, err
			}

			files = append(files, filepath.Join(c.fi.Path, name))
		}

	} else {
		files = append(files, c.fi.Path)
	}

	if query == "true" {
		query = "zip"
	}

	var (
		extension string
		temp      string
		err       error
		tempfile  string
	)

	temp, err = ioutil.TempDir("", "")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	defer os.RemoveAll(temp)
	tempfile = filepath.Join(temp, "temp")

	switch query {
	case "zip":
		extension, err = ".zip", archiver.Zip.Make(tempfile, files)
	case "tar":
		extension, err = ".tar", archiver.Tar.Make(tempfile, files)
	case "targz":
		extension, err = ".tar.gz", archiver.TarGz.Make(tempfile, files)
	case "tarbz2":
		extension, err = ".tar.bz2", archiver.TarBz2.Make(tempfile, files)
	case "tarxz":
		extension, err = ".tar.xz", archiver.TarXZ.Make(tempfile, files)
	default:
		return http.StatusNotImplemented, nil
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	file, err := os.Open(temp + "/temp")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	name := c.fi.Name
	if name == "." || name == "" {
		name = "download"
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+name+extension)
	io.Copy(w, file)
	return 0, nil
}

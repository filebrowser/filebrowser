package handlers

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/file"
	"github.com/mholt/archiver"
)

// Download creates an archive in one of the supported formats (zip, tar,
// tar.gz or tar.bz2) and sends it to be downloaded.
func Download(w http.ResponseWriter, r *http.Request, c *config.Config, i *file.Info) (int, error) {
	query := r.URL.Query().Get("download")

	if !i.IsDir {
		w.Header().Set("Content-Disposition", "attachment; filename="+i.Name)
		http.ServeFile(w, r, i.Path)
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

			files = append(files, filepath.Join(i.Path, name))
		}

	} else {
		files = append(files, i.Path)
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

	name := i.Name
	if name == "." || name == "" {
		name = "download"
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+name+extension)
	io.Copy(w, file)
	return http.StatusOK, nil
}

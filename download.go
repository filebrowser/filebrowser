package filemanager

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/filemanager/dir"
	"github.com/mholt/archiver"
)

// downloadHandler creates an archive in one of the supported formats (zip, tar,
// tar.gz or tar.bz2) and sends it to be downloaded.
func downloadHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	query := r.URL.Query().Get("format")

	// If the file isn't a directory, serve it using http.ServeFile. We display it
	// inline if it is requested.
	if !c.FI.IsDir {
		if r.URL.Query().Get("inline") == "true" {
			w.Header().Set("Content-Disposition", "inline")
		} else {
			w.Header().Set("Content-Disposition", "attachment; filename="+c.FI.Name)
		}

		http.ServeFile(w, r, c.FI.Path)
		return 0, nil
	}

	files := []string{}
	names := strings.Split(r.URL.Query().Get("files"), ",")

	// If there are files in the query, sanitize their names.
	// Otherwise, just append the current path.
	if len(names) != 0 {
		for _, name := range names {
			// Unescape the name.
			name, err := url.QueryUnescape(name)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			// Clean the slashes.
			name = dir.SlashClean(name)
			files = append(files, filepath.Join(c.FI.Path, name))
		}
	} else {
		files = append(files, c.FI.Path)
	}

	// If the format is true, just set it to "zip".
	if query == "true" {
		query = "zip"
	}

	var (
		extension string
		temp      string
		err       error
		tempfile  string
	)

	// Create a temporary directory.
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

	// Defines the file name.
	name := c.FI.Name
	if name == "." || name == "" {
		name = "download"
	}
	name += extension

	// Opens the file so it can be downloaded.
	file, err := os.Open(temp + "/temp")
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+name)
	_, err = io.Copy(w, file)
	return 0, err
}

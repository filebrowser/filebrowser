package http

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	fb "github.com/filebrowser/filebrowser/lib"
	"github.com/hacdias/fileutils"
	"github.com/mholt/archiver"
)

// downloadHandler creates an archive in one of the supported formats (zip, tar,
// tar.gz or tar.bz2) and sends it to be downloaded.
func downloadHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	// If the file isn't a directory, serve it using http.ServeFile. We display it
	// inline if it is requested.
	if !c.File.IsDir {
		return downloadFileHandler(c, w, r)
	}

	query := r.URL.Query().Get("format")
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
			name = fileutils.SlashClean(name)
			files = append(files, filepath.Join(c.File.Path, name))
		}
	} else {
		files = append(files, c.File.Path)
	}

	var (
		extension string
		ar        archiver.Archiver
	)

	switch query {
	// If the format is true, just set it to "zip".
	case "zip", "true", "":
		extension, ar = ".zip", archiver.Zip
	case "tar":
		extension, ar = ".tar", archiver.Tar
	case "targz":
		extension, ar = ".tar.gz", archiver.TarGz
	case "tarbz2":
		extension, ar = ".tar.bz2", archiver.TarBz2
	case "tarxz":
		extension, ar = ".tar.xz", archiver.TarXZ
	default:
		return http.StatusNotImplemented, nil
	}

	// Defines the file name.
	name := c.File.Name
	if name == "." || name == "" {
		name = "archive"
	}
	name += extension

	w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(name))
	err := ar.Write(w, files)

	return 0, err
}

func downloadFileHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	file, err := os.Open(c.File.Path)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if r.URL.Query().Get("inline") == "true" {
		w.Header().Set("Content-Disposition", "inline")
	} else {
		// As per RFC6266 section 4.3
		w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(c.File.Name))
	}

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)

	return 0, nil
}

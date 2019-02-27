package http

import (
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/hacdias/fileutils"
	"github.com/mholt/archiver"
)

func parseQueryFiles(r *http.Request, f *files.FileInfo, u *users.User) ([]string, error) {
	files := []string{}
	names := strings.Split(r.URL.Query().Get("files"), ",")

	if len(names) == 0 {
		files = append(files, f.Path)
	} else {
		for _, name := range names {
			name, err := url.QueryUnescape(strings.Replace(name, "+", "%2B", -1))
			if err != nil {
				return nil, err
			}

			name = fileutils.SlashClean(name)
			files = append(files, filepath.Join(f.Path, name))
		}
	}

	return files, nil
}

func parseQueryAlgorithm(r *http.Request) (string, archiver.Writer, error) {
	switch r.URL.Query().Get("algo") {
	case "zip", "true", "":
		return ".zip", archiver.NewZip(), nil
	case "tar":
		return ".tar", archiver.NewTar(), nil
	case "targz":
		return ".tar.gz", archiver.NewTarGz(), nil
	case "tarbz2":
		return ".tar.bz2", archiver.NewTarBz2(), nil
	case "tarxz":
		return ".tar.xz", archiver.NewTarXz(), nil
	case "tarlz4":
		return ".tar.lz4", archiver.NewTarLz4(), nil
	case "tarsz":
		return ".tar.sz", archiver.NewTarSz(), nil
	default:
		return "", nil, errors.New("format not implemented")
	}
}

var rawHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusAccepted, nil
	}

	file, err := files.NewFileInfo(files.FileOptions{
		Fs:      d.user.Fs,
		Path:    r.URL.Path,
		Modify:  d.user.Perm.Modify,
		Expand:  false,
		Checker: d,
	})
	if err != nil {
		return errToStatus(err), err
	}

	if !file.IsDir {
		return rawFileHandler(w, r, file)
	}

	return rawDirHandler(w, r, d, file)
})

func addFile(ar archiver.Writer, d *data, path string) error {
	// Checks are always done with paths with "/" as path separator.
	path = strings.Replace(path, "\\", "/", -1)
	if !d.Check(path) {
		return nil
	}

	info, err := d.user.Fs.Stat(path)
	if err != nil {
		return err
	}

	file, err := d.user.Fs.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = ar.Write(archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo:   info,
			CustomName: strings.TrimPrefix(path, "/"),
		},
		ReadCloser: file,
	})
	if err != nil {
		return err
	}

	if info.IsDir() {
		names, err := file.Readdirnames(0)
		if err != nil {
			return err
		}

		for _, name := range names {
			err = addFile(ar, d, filepath.Join(path, name))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func rawDirHandler(w http.ResponseWriter, r *http.Request, d *data, file *files.FileInfo) (int, error) {
	filenames, err := parseQueryFiles(r, file, d.user)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	extension, ar, err := parseQueryAlgorithm(r)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	name := file.Name
	if name == "." || name == "" {
		name = "archive"
	}
	name += extension
	w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(name))

	err = ar.Create(w)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer ar.Close()

	for _, fname := range filenames {
		err = addFile(ar, d, fname)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	return 0, nil
}

func rawFileHandler(w http.ResponseWriter, r *http.Request, file *files.FileInfo) (int, error) {
	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer fd.Close()

	if r.URL.Query().Get("inline") == "true" {
		w.Header().Set("Content-Disposition", "inline")
	} else {
		// As per RFC6266 section 4.3
		w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(file.Name))
	}

	http.ServeContent(w, r, file.Name, file.ModTime, fd)
	return 0, nil
}

package http

import (
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/filebrowser/filebrowser/lib"
	"github.com/hacdias/fileutils"
	"github.com/mholt/archiver"
)

const apiRawPrefix = "/api/raw"

func parseQueryFiles(r *http.Request, f *lib.File, u *lib.User) ([]string, error) {
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

func (e *Env) rawHandler(w http.ResponseWriter, r *http.Request) {
	path, user, ok := e.getResourceData(w, r, apiRawPrefix)
	if !ok {
		return
	}

	if !user.Perm.Download {
		httpErr(w, r, http.StatusForbidden, nil)
		return
	}

	file, err := e.NewFile(path, user)
	if err != nil {
		httpErr(w, r, httpFsErr(err), err)
		return
	}

	if !file.IsDir {
		fileHandler(w, r, file, user)
		return
	}

	filenames, err := parseQueryFiles(r, file, user)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	extension, ar, err := parseQueryAlgorithm(r)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	name := file.Name
	if name == "." || name == "" {
		name = "archive"
	}
	name += extension
	w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(name))

	err = ar.Create(w)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}
	defer ar.Close()

	for _, fname := range filenames {
		info, err := user.Fs.Stat(fname)
		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
			return
		}

		// get file's name for the inside of the archive
		internalName, err := archiver.NameInArchive(info, fname, fname)
		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
			return
		}

		// open the file
		file, err := user.Fs.Open(fname)
		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
			return
		}

		// write it to the archive
		err = ar.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   info,
				CustomName: internalName,
			},
			ReadCloser: file,
		})
		file.Close()
		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
			return
		}
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request, file *lib.File, user *lib.User) {
	fd, err := user.Fs.Open(file.Path)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}
	defer fd.Close()

	if r.URL.Query().Get("inline") == "true" {
		w.Header().Set("Content-Disposition", "inline")
	} else {
		// As per RFC6266 section 4.3
		w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(file.Name))
	}

	http.ServeContent(w, r, file.Name, file.ModTime, fd)
}

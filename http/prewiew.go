package http

import (
	"fmt"
	"github.com/gorilla/mux"
	"image"
	"net/http"
	"net/url"
	"strings"

	"github.com/disintegration/imaging"

	"github.com/filebrowser/filebrowser/v2/files"
)

const (
	sizeThumb = "thumb"
	sizeBig   = "big"
)

type imageProcessor func(src image.Image) (image.Image, error)

var previewHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusAccepted, nil
	}
	vars := mux.Vars(r)
	size := vars["size"]
	if size != sizeBig && size != sizeThumb {
		return http.StatusNotImplemented, nil
	}

	// Resolve file path from URL
	path := "/" + strings.Join(strings.Split(r.URL.Path, "/")[2:], "/")
	file, err := files.NewFileInfo(files.FileOptions{
		Fs:      d.user.Fs,
		Path:    path,
		Modify:  d.user.Perm.Modify,
		Expand:  true,
		Checker: d,
	})
	if err != nil {
		return errToStatus(err), err
	}

	if r.URL.Query().Get("inline") == "true" {
		w.Header().Set("Content-Disposition", "inline")
	} else {
		// As per RFC6266 section 4.3
		w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(file.Name))
	}

	switch file.Type {
	case "image":
		return handleImagePreview(w, file, size)
	default:
		return http.StatusNotImplemented, fmt.Errorf("can't create preview for %s type", file.Type)
	}
})

func handleImagePreview(w http.ResponseWriter, file *files.FileInfo, size string) (int, error) {
	var imgProcessor imageProcessor
	switch size {
	case sizeBig:
		imgProcessor = func(img image.Image) (image.Image, error) {
			return imaging.Fit(img, 1080, 1080, imaging.Lanczos), nil
		}
	case sizeThumb:
		imgProcessor = func(img image.Image) (image.Image, error) {
			return imaging.Thumbnail(img, 128, 128, imaging.Box), nil
		}
	default:
		return http.StatusBadRequest, fmt.Errorf("unsupported preview size %s", size)
	}

	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return errToStatus(err), err
	}
	defer fd.Close()
	format, err := imaging.FormatFromExtension(file.Extension)
	if err != nil {
		return http.StatusNotImplemented, err
	}
	img, err := imaging.Decode(fd, imaging.AutoOrientation(true))
	if err != nil {
		return errToStatus(err), err
	}
	img, err = imgProcessor(img)
	if err != nil {
		return errToStatus(err), err
	}
	if imaging.Encode(w, img, format) != nil {
		return errToStatus(err), err
	}
	return 0, nil
}

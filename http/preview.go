package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/img"
)

const (
	sizeThumb = "thumb"
	sizeBig   = "big"
)

type ImgService interface {
	FormatFromExtension(ext string) (img.Format, error)
	Resize(ctx context.Context, in io.Reader, width, height int, out io.Writer, options ...img.Option) error
}

func previewHandler(imgSvc ImgService, enableThumbnails, resizePreview bool) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Download {
			return http.StatusAccepted, nil
		}
		vars := mux.Vars(r)
		size := vars["size"]
		if size != sizeBig && size != sizeThumb {
			return http.StatusNotImplemented, nil
		}

		file, err := files.NewFileInfo(files.FileOptions{
			Fs:      d.user.Fs,
			Path:    "/" + vars["path"],
			Modify:  d.user.Perm.Modify,
			Expand:  true,
			Checker: d,
		})
		if err != nil {
			return errToStatus(err), err
		}

		setContentDisposition(w, r, file)

		switch file.Type {
		case "image":
			return handleImagePreview(imgSvc, w, r, file, size, enableThumbnails, resizePreview)
		default:
			return http.StatusNotImplemented, fmt.Errorf("can't create preview for %s type", file.Type)
		}
	})
}

func handleImagePreview(imgSvc ImgService, w http.ResponseWriter, r *http.Request,
	file *files.FileInfo, size string, enableThumbnails, resizePreview bool) (int, error) {
	format, err := imgSvc.FormatFromExtension(file.Extension)
	if err != nil {
		// Unsupported extensions directly return the raw data
		if err == img.ErrUnsupportedFormat {
			return rawFileHandler(w, r, file)
		}
		return errToStatus(err), err
	}

	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return errToStatus(err), err
	}
	defer fd.Close()

	var (
		width   int
		height  int
		options []img.Option
	)

	switch {
	case size == sizeBig && resizePreview && format != img.FormatGif:
		width = 1080
		height = 1080
		options = append(options, img.WithMode(img.ResizeModeFit), img.WithQuality(img.QualityMedium))
	case size == sizeThumb && enableThumbnails:
		width = 128
		height = 128
		options = append(options, img.WithMode(img.ResizeModeFill), img.WithQuality(img.QualityLow), img.WithFormat(img.FormatJpeg))
	default:
		if _, err := rawFileHandler(w, r, file); err != nil {
			return errToStatus(err), err
		}
		return 0, nil
	}

	if err := imgSvc.Resize(r.Context(), fd, width, height, w, options...); err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		default:
			return 0, err
		}
	}
	return 0, nil
}

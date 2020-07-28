//go:generate go-enum --sql --marshal --names --file $GOFILE
package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/img"
)

/*
ENUM(
thumb
big
)
*/
type PreviewSize int

type ImgService interface {
	FormatFromExtension(ext string) (img.Format, error)
	Resize(ctx context.Context, in io.Reader, width, height int, out io.Writer, options ...img.Option) error
}

type FileCache interface {
	Store(ctx context.Context, key string, value []byte) error
	Load(ctx context.Context, key string) ([]byte, bool, error)
	Delete(ctx context.Context, key string) error
}

func previewHandler(imgSvc ImgService, fileCache FileCache, enableThumbnails, resizePreview bool) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Download {
			return http.StatusAccepted, nil
		}
		vars := mux.Vars(r)

		previewSize, err := ParsePreviewSize(vars["size"])
		if err != nil {
			return http.StatusBadRequest, err
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
			return handleImagePreview(w, r, imgSvc, fileCache, file, previewSize, enableThumbnails, resizePreview)
		default:
			return http.StatusNotImplemented, fmt.Errorf("can't create preview for %s type", file.Type)
		}
	})
}

func handleImagePreview(w http.ResponseWriter, r *http.Request, imgSvc ImgService, fileCache FileCache,
	file *files.FileInfo, previewSize PreviewSize, enableThumbnails, resizePreview bool) (int, error) {
	format, err := imgSvc.FormatFromExtension(file.Extension)
	if err != nil {
		// Unsupported extensions directly return the raw data
		if err == img.ErrUnsupportedFormat {
			return rawFileHandler(w, r, file)
		}
		return errToStatus(err), err
	}

	cacheKey := previewCacheKey(file.Path, previewSize)
	cachedFile, ok, err := fileCache.Load(r.Context(), cacheKey)
	if err != nil {
		return errToStatus(err), err
	}
	if ok {
		_, _ = w.Write(cachedFile)
		return 0, nil
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
	case previewSize == PreviewSizeBig && resizePreview && format != img.FormatGif:
		width = 1080
		height = 1080
		options = append(options, img.WithMode(img.ResizeModeFit), img.WithQuality(img.QualityMedium))
	case previewSize == PreviewSizeThumb && enableThumbnails:
		width = 128
		height = 128
		options = append(options, img.WithMode(img.ResizeModeFill), img.WithQuality(img.QualityLow), img.WithFormat(img.FormatJpeg))
	default:
		if _, err := rawFileHandler(w, r, file); err != nil {
			return errToStatus(err), err
		}
		return 0, nil
	}

	buf := &bytes.Buffer{}
	if err := imgSvc.Resize(context.Background(), fd, width, height, buf, options...); err != nil {
		return 0, err
	}

	go func() {
		if err := fileCache.Store(context.Background(), cacheKey, buf.Bytes()); err != nil {
			fmt.Printf("failed to cache resized image: %v", err)
		}
	}()

	_, _ = w.Write(buf.Bytes())

	return 0, nil
}

func previewCacheKey(fPath string, previewSize PreviewSize) string {
	return fPath + previewSize.String()
}

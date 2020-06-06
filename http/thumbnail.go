package http

import (
	"github.com/disintegration/imaging"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"mime"
	"net/http"
	"net/url"
)

var thumbnailHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusAccepted, nil
	}

	file, err := files.NewFileInfo(files.FileOptions{
		Fs:      d.user.Fs,
		Path:    r.URL.Path,
		Modify:  d.user.Perm.Modify,
		Expand:  true,
		Checker: d,
	})
	if err != nil {
		return errToStatus(err), err
	}

	if file.IsDir || file.Type != "image" {
		return http.StatusNotFound, nil
	}

	return thumbnailFileHandler(w, r, file)
})

func thumbnailFileHandler(w http.ResponseWriter, r *http.Request, file *files.FileInfo) (int, error) {
	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return errToStatus(err), err
	}
	defer fd.Close()

	if r.URL.Query().Get("inline") == "true" {
		w.Header().Set("Content-Disposition", "inline")
	} else {
		// As per RFC6266 section 4.3
		w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(file.Name))
	}

	srcImg, err := imaging.Decode(fd, imaging.AutoOrientation(true))
	if err != nil {
		return errToStatus(err), err
	}
	dstImg := fitResizeImage(srcImg)
	w.Header().Add("Content-Type", mime.TypeByExtension(file.Extension))
	err = func() error {
		switch file.Extension {
		case ".jpg", ".jpeg":
			return jpeg.Encode(w, dstImg, nil)
		case ".png":
			return png.Encode(w, dstImg)
		case ".gif":
			return gif.Encode(w, dstImg, nil)
		default:
			return errors.ErrNotExist
		}
	}()
	if err != nil {
		return errToStatus(err), err
	}
	return 0, nil
}

const maxSize = 1080

func fitResizeImage(srcImage image.Image) image.Image {
	width := srcImage.Bounds().Dx()
	height := srcImage.Bounds().Dy()
	if width > maxSize && width > height {
		width = maxSize
		height = 0
	} else if height > maxSize && height > width {
		width = 0
		height = maxSize
	} else {
		return srcImage
	}
	return imaging.Resize(srcImage, width, height, imaging.NearestNeighbor)
}

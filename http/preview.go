package http

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/nfnt/resize"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime"
	"net/http"
	"net/url"
)

var compressHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
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

	return compressFileHandler(w, r, file)
})

func compressFileHandler(w http.ResponseWriter, r *http.Request, file *files.FileInfo) (int, error) {
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

	buf, err := compressImageHandler(file, fd)
	if err != nil {
		return errToStatus(err), err
	}
	w.Header().Add("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.Header().Add("Content-Type", mime.TypeByExtension(file.Extension))
	io.Copy(w, buf)
	return 0, nil
}

func compressImageHandler(file *files.FileInfo, fd io.Reader) (*bytes.Buffer, error) {
	var (
		buf *bytes.Buffer
		m   image.Image
		err error
	)

	switch file.Extension {
	case ".jpg", ".jpeg":
		buf, m, err = compressImage(jpeg.Decode, fd)
		if err != nil {
			return nil, err
		}
		err = jpeg.Encode(buf, m, nil)
		break
	case ".png":
		buf, m, err = compressImage(png.Decode, fd)
		if err != nil {
			return nil, err
		}
		err = png.Encode(buf, m)
		break
	case ".gif":
		buf, m, err = compressImage(gif.Decode, fd)
		if err != nil {
			return nil, err
		}
		err = gif.Encode(buf, m, nil)
		break
	default:
		return nil, errors.New("extension is not supported")
	}
	if err != nil {
		return nil, err
	}
	return buf, nil
}

const maxSize = 1080

func compressImage(decode func(r io.Reader) (image.Image, error), fd io.Reader) (*bytes.Buffer, image.Image, error) {
	img, err := decode(fd)
	if err != nil {
		return nil, nil, err
	}
	buf := bytes.NewBuffer([]byte{})
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	if width > maxSize && width > height {
		width = maxSize
		height = 0
	} else if height > maxSize && height > width {
		width = 0
		height = maxSize
	} else {
		width = 0
		height = 0
	}
	m := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)
	return buf, m, nil
}

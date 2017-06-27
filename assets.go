package filemanager

import (
	"errors"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// assetsURL is the url where static assets are served.
const assetsURL = "/_internal"

// Serve provides the needed assets for the front-end
func serveAssets(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// gets the filename to be used with Assets function
	filename := strings.TrimPrefix(r.URL.Path, assetsURL)

	var file []byte
	var err error

	switch {
	case strings.HasPrefix(filename, "/css"):
		filename = strings.Replace(filename, "/css/", "", 1)
		file, err = c.fm.assets.css.Bytes(filename)
	case strings.HasPrefix(filename, "/js"):
		filename = strings.Replace(filename, "/js/", "", 1)
		file, err = c.fm.assets.js.Bytes(filename)
	default:
		err = errors.New("not found")
	}

	if err != nil {
		return http.StatusNotFound, nil
	}

	// Get the file extension and its mimetype
	extension := filepath.Ext(filename)
	mediatype := mime.TypeByExtension(extension)

	// Write the header with the Content-Type and write the file
	// content to the buffer
	w.Header().Set("Content-Type", mediatype)
	w.Write(file)
	return 200, nil
}

package filemanager

import (
	"errors"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// AssetsURL is the url of the assets
const AssetsURL = "/_internal"

// Serve provides the needed assets for the front-end
func serveAssets(w http.ResponseWriter, r *http.Request, m *FileManager) (int, error) {
	// gets the filename to be used with Assets function
	filename := strings.Replace(r.URL.Path, m.BaseURL+AssetsURL, "", 1)

	var file []byte
	var err error

	switch {
	case strings.HasPrefix(filename, "/css"):
		filename = strings.Replace(filename, "/css/", "", 1)
		file, err = m.Assets.CSS.Bytes(filename)
	case strings.HasPrefix(filename, "/js"):
		filename = strings.Replace(filename, "/js/", "", 1)
		file, err = m.Assets.requiredJS.Bytes(filename)
	case strings.HasPrefix(filename, "/vendor"):
		if m.Assets.JS != nil {
			filename = strings.Replace(filename, "/vendor/", "", 1)
			file, err = m.Assets.JS.Bytes(filename)
			break
		}

		fallthrough
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

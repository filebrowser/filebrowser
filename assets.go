package filemanager

import (
	"errors"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// ServeAssets redirects the request for the respective method
func ServeAssets(w http.ResponseWriter, r *http.Request, c *Config) (int, error) {
	switch r.Method {
	case "GET":
		return serveAssetsGET(w, r, c)
	default:
		return http.StatusMethodNotAllowed, errors.New("Invalid method.")
	}
}

// serveAssetsGET provides the method for GET request on Assets page
func serveAssetsGET(w http.ResponseWriter, r *http.Request, c *Config) (int, error) {
	// gets the filename to be used with Assets function
	filename := strings.Replace(r.URL.Path, c.BaseURL+"/_filemanagerinternal", "public", 1)
	file, err := Asset(filename)
	if err != nil {
		return 404, nil
	}

	// Get the file extension ant its mime type
	extension := filepath.Ext(filename)
	mediatype := mime.TypeByExtension(extension)

	// Write the header with the Content-Type and write the file
	// content to the buffer
	w.Header().Set("Content-Type", mediatype)
	w.Write(file)
	return 200, nil
}

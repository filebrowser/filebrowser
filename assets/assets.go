package assets

import (
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/config"
)

// BaseURL is the url of the assets
const BaseURL = "/_filemanagerinternal"

// Serve provides the needed assets for the front-end
func Serve(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// gets the filename to be used with Assets function
	filename := strings.Replace(r.URL.Path, c.BaseURL+BaseURL, "public", 1)
	file, err := Asset(filename)
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

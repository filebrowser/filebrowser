package browse

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/hacdias/caddy-hugo/tools/server"
)

// PUT handles the HTTP PUT request for all /admin/browse related requests.
// Renames a file and/or a folder.
func PUT(w http.ResponseWriter, r *http.Request) (int, error) {
	// Remove both beginning and trailing slashes
	old := r.URL.Path
	old = strings.TrimPrefix(old, "/")
	old = strings.TrimSuffix(old, "/")
	old = conf.Path + old

	// Get the JSON information sent using a buffer
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(r.Body)

	// Creates the raw file "map" using the JSON
	var info map[string]interface{}
	json.Unmarshal(buffer.Bytes(), &info)

	// Check if filename and archetype are specified in
	// the request
	if _, ok := info["filename"]; !ok {
		return server.RespondJSON(w, &response{"Filename not specified.", ""}, http.StatusBadRequest, nil)
	}

	// Sanitize the file name path
	new := info["filename"].(string)
	new = strings.TrimPrefix(new, "/")
	new = strings.TrimSuffix(new, "/")
	new = conf.Path + new

	// Renames the file/folder
	if err := os.Rename(old, new); err != nil {
		return server.RespondJSON(w, &response{err.Error(), ""}, http.StatusInternalServerError, err)
	}

	return server.RespondJSON(w, &response{"File renamed.", ""}, http.StatusOK, nil)
}

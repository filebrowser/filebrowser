package browse

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/tools/utils"
)

// PUT handles the HTTP PUT request for all /admin/browse related requests.
// Renames a file and/or a folder.
func PUT(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Remove both beginning and trailing slashes
	old := r.URL.Path
	old = strings.TrimPrefix(old, "/")
	old = strings.TrimSuffix(old, "/")
	old = c.Path + old

	// Get the JSON information sent using a buffer
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(r.Body)

	// Creates the raw file "map" using the JSON
	var info map[string]interface{}
	json.Unmarshal(buffer.Bytes(), &info)

	// Check if filename and archetype are specified in
	// the request
	if _, ok := info["filename"]; !ok {
		return utils.RespondJSON(w, map[string]string{
			"message": "Filename not specified.",
		}, 400, nil)
	}

	// Sanitize the file name path
	new := info["filename"].(string)
	new = strings.TrimPrefix(new, "/")
	new = strings.TrimSuffix(new, "/")
	new = c.Path + new

	// Renames the file/folder
	if err := os.Rename(old, new); err != nil {
		return utils.RespondJSON(w, map[string]string{
			"message": "Something went wrong.",
		}, 500, err)
	}

	return utils.RespondJSON(w, map[string]string{
		"message": "File renamed.",
	}, 200, nil)
}

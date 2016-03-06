package browse

import (
	"net/http"
	"os"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/utils"
)

// DELETE handles the delete requests on browse pages
func DELETE(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Remove both beginning and trailing slashes
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	path = c.Path + path

	message := "File deleted."

	// Check if the file or directory exists
	if stat, err := os.Stat(path); err == nil {
		var err error
		// If it's dir, remove all of the content inside
		if stat.IsDir() {
			err = os.RemoveAll(path)
			message = "Folder deleted."
		} else {
			err = os.Remove(path)
		}

		// Check for errors
		if err != nil {
			return utils.RespondJSON(w, "Something went wrong.", 500, nil)
		}
	} else {
		return utils.RespondJSON(w, "File not found.", 404, nil)
	}

	return utils.RespondJSON(w, message, 200, nil)
}

package browse

import (
	"net/http"
	"os"
	"strings"

	s "github.com/hacdias/caddy-hugo/tools/server"
)

// DELETE handles the delete requests on browse pages
func DELETE(w http.ResponseWriter, r *http.Request) (int, error) {
	// Remove both beginning and trailing slashes
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	path = conf.Path + path

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
			return s.RespondJSON(w, &response{err.Error(), ""}, http.StatusInternalServerError, err)
		}
	} else {
		return s.RespondJSON(w, &response{"File not found.", ""}, http.StatusNotFound, nil)
	}

	return s.RespondJSON(w, &response{message, ""}, http.StatusOK, nil)
}

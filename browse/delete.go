package browse

import (
	"net/http"
	"os"
	"strings"
)

// DELETE handles the DELETE method on browse page
func DELETE(w http.ResponseWriter, r *http.Request) (int, error) {
	// Remove both beginning and trailing slashes
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/")
	r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")

	// Check if the file or directory exists
	if stat, err := os.Stat(r.URL.Path); err == nil {
		var err error
		// If it's dir, remove all of the content inside
		if stat.IsDir() {
			err = os.RemoveAll(r.URL.Path)
		} else {
			err = os.Remove(r.URL.Path)
		}

		// Check for errors
		if err != nil {
			w.Write([]byte(err.Error()))
			return 500, err
		}
	} else {
		return 404, nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
	return 200, nil
}

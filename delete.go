package filemanager

import (
	"net/http"
	"os"
)

// Delete handles the delete requests
func Delete(path string, info os.FileInfo) (int, error) {
	var err error
	// If it's dir, remove all of the content inside
	if info.IsDir() {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}

	// Check for errors
	if err != nil {
		switch {
		case os.IsPermission(err):
			return http.StatusForbidden, err
		case os.IsExist(err):
			return http.StatusGone, err
		default:
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusOK, nil
}

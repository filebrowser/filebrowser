package errors

import (
	"net/http"
	"os"
)

// ToHTTPCode gets the respective HTTP code for an error
func ToHTTPCode(err error) int {
	switch {
	case os.IsPermission(err):
		return http.StatusForbidden
	case os.IsNotExist(err):
		return http.StatusNotFound
	case os.IsExist(err):
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}

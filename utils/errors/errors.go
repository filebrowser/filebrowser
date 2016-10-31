package errors

import (
	"net/http"
	"os"
)

// ErrorToHTTPCode converts errors to HTTP Status Code.
func ErrorToHTTPCode(err error, gone bool) int {
	switch {
	case os.IsPermission(err):
		return http.StatusForbidden
	case os.IsNotExist(err):
		if !gone {
			return http.StatusNotFound
		}

		return http.StatusGone
	case os.IsExist(err):
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}

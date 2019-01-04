package http

import (
	"encoding/json"
	"net/http"
	"os"
)

func renderJSON(w http.ResponseWriter, r *http.Request, data interface{}) (int, error) {
	marsh, err := json.Marshal(data)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(marsh); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

func errToStatus(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case os.IsPermission(err):
		return http.StatusForbidden
	case os.IsNotExist(err):
		return http.StatusNotFound
	case os.IsExist(err):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

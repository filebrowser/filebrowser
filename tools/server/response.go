package server

import (
	"encoding/json"
	"errors"
	"net/http"
)

// RespondJSON
func RespondJSON(w http.ResponseWriter, message map[string]string, code int, err error) (int, error) {
	msg, msgErr := json.Marshal(message)

	if msgErr != nil {
		return 500, msgErr
	}

	if code == 500 && err != nil {
		err = errors.New(message["message"])
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(msg)
	return 0, err
}

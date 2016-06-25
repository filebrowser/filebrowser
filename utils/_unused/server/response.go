package server

import (
	"encoding/json"
	"net/http"
)

// RespondJSON used to send JSON responses to the web server
func RespondJSON(w http.ResponseWriter, message interface{}, code int, err error) (int, error) {
	if message == nil {
		message = map[string]string{}
	}

	msg, msgErr := json.Marshal(message)

	if msgErr != nil {
		return 500, msgErr
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(msg)
	return 0, err
}

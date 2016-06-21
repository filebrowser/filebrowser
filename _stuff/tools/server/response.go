package server

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code    int
	Err     error
	Content interface{}
}

// RespondJSON used to send JSON responses to the web server
func RespondJSON(w http.ResponseWriter, r *Response) (int, error) {
	if r.Content == nil {
		r.Content = map[string]string{}
	}

	msg, msgErr := json.Marshal(r.Content)

	if msgErr != nil {
		return 500, msgErr
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)
	w.Write(msg)
	return 0, r.Err
}

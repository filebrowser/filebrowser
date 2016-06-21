package hugo

import (
	"encoding/json"
	"net/http"
)

// Response conta
type Response struct {
	Code    int
	Err     error
	Content string
}

// Send used to send JSON responses to the web server
func (r *Response) Send(w http.ResponseWriter) (int, error) {
	content := map[string]string{"message": r.Content}
	msg, msgErr := json.Marshal(content)

	if msgErr != nil {
		return 500, msgErr
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)
	w.Write(msg)
	return 0, r.Err
}

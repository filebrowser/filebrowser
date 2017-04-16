package wrapper

import "net/http"

// ResponseWriterNoBody is a wrapper used to suprress the body of the response
// to a request. Mainly used for HEAD requests.
type ResponseWriterNoBody struct {
	http.ResponseWriter
}

// NewResponseWriterNoBody creates a new ResponseWriterNoBody.
func NewResponseWriterNoBody(w http.ResponseWriter) *ResponseWriterNoBody {
	return &ResponseWriterNoBody{w}
}

// Header executes the Header method from the http.ResponseWriter.
func (w ResponseWriterNoBody) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Write suprresses the body.
func (w ResponseWriterNoBody) Write(data []byte) (int, error) {
	return 0, nil
}

// WriteHeader writes the header to the http.ResponseWriter.
func (w ResponseWriterNoBody) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

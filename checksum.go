package filemanager

import "net/http"

// serveChecksum calculates the hash of a file. Supports MD5, SHA1, SHA256 and SHA512.
func serveChecksum(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	query := r.URL.Query().Get("checksum")

	val, err := c.fi.Checksum(query)
	if err == errInvalidOption {
		return http.StatusBadRequest, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Write([]byte(val))
	return http.StatusOK, nil
}

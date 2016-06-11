package filemanager

import "net/http"

// ServeSingleFile redirects the request for the respective method
func (f FileManager) ServeSingleFile(w http.ResponseWriter, r *http.Request, file http.File, c *Config) (int, error) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Hello"))
	return 200, nil
}

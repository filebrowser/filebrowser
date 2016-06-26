package file

import (
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/internal/config"
)

// Upload is used to handle the upload requests to the server
func Upload(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Parse the multipart form in the request
	err := r.ParseMultipartForm(100000)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError, err
	}

	// For each file header in the multipart form
	for _, headers := range r.MultipartForm.File {
		// Handle each file
		for _, header := range headers {
			// Open the first file
			var src multipart.File
			if src, err = header.Open(); nil != err {
				return http.StatusInternalServerError, err
			}

			filename := strings.Replace(r.URL.Path, c.BaseURL, c.PathScope, 1)
			filename = filename + header.Filename
			filename = filepath.Clean(filename)

			// Create the file
			var dst *os.File
			if dst, err = os.Create(filename); nil != err {
				if os.IsExist(err) {
					return http.StatusConflict, err
				}
				return http.StatusInternalServerError, err
			}

			// Copy the file content
			if _, err = io.Copy(dst, src); nil != err {
				return http.StatusInternalServerError, err
			}

			defer dst.Close()
		}
	}

	return http.StatusOK, nil
}

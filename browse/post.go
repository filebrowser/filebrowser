package browse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/utils"
)

// POST handles the POST method on browse page. It's used to create new files,
// folders and upload content.
func POST(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Remove prefix slash
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/")

	// If it's the upload of a file
	if r.Header.Get("X-Upload") == "true" {
		return upload(w, r, c)
	}

	// Get the JSON information sent using a buffer
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(r.Body)

	// Creates the raw file "map" using the JSON
	var info map[string]interface{}
	json.Unmarshal(buffer.Bytes(), &info)

	// Check if filename and archetype are specified in
	// the request
	if _, ok := info["filename"]; !ok {
		return http.StatusBadRequest, errors.New("Filename not specified.")
	}

	if _, ok := info["archetype"]; !ok {
		return http.StatusBadRequest, errors.New("Archtype not specified.")
	}

	// Sanitize the file name path
	filename := info["filename"].(string)
	filename = strings.TrimPrefix(filename, "/")
	filename = strings.TrimSuffix(filename, "/")
	url := "/admin/edit/" + r.URL.Path + filename
	filename = c.Path + r.URL.Path + filename

	if strings.HasPrefix(filename, c.Path+"content/") &&
		(strings.HasSuffix(filename, ".md") || strings.HasSuffix(filename, ".markdown")) {

		filename = strings.Replace(filename, c.Path+"content/", "", 1)
		args := []string{"new", filename}
		archetype := info["archetype"].(string)

		if archetype != "" {
			args = append(args, "--kind", archetype)
		}

		if err := utils.RunCommand(c.Hugo, args, c.Path); err != nil {
			return http.StatusInternalServerError, err
		}
	} else {
		var err error

		if filepath.Ext(filename) == "" {
			err = os.MkdirAll(filename, 0755)
			url = strings.Replace(url, "edit", "browse", 1)
		} else {
			var wf *os.File
			wf, err = os.Create(filename)
			defer wf.Close()
		}

		if err != nil {
			return http.StatusInternalServerError, err
		}

	}

	fmt.Println(url)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Location\": \"" + url + "\"}"))
	return http.StatusOK, nil
}

func upload(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Parse the multipart form in the request
	err := r.ParseMultipartForm(100000)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// For each file header in the multipart form
	for _, fheaders := range r.MultipartForm.File {
		// Handle each file
		for _, hdr := range fheaders {
			// Open the first file
			var infile multipart.File
			if infile, err = hdr.Open(); nil != err {
				return http.StatusInternalServerError, err
			}

			// Create the file
			var outfile *os.File
			if outfile, err = os.Create(c.Path + r.URL.Path + hdr.Filename); nil != err {
				return http.StatusInternalServerError, err
			}

			// Copy the file content
			if _, err = io.Copy(outfile, infile); nil != err {
				return http.StatusInternalServerError, err
			}

			defer outfile.Close()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
	return http.StatusOK, nil
}

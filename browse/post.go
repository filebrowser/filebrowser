package browse

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/utils"
)

// POST handles the POST method on browse page
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
	filename = c.Path + r.URL.Path + filename

	// Check if the archetype is defined
	if info["archetype"] != "" {
		// Sanitize the archetype path
		archetype := info["archetype"].(string)
		archetype = strings.Replace(archetype, "/archetypes", "", 1)
		archetype = strings.Replace(archetype, "archetypes", "", 1)
		archetype = strings.TrimPrefix(archetype, "/")
		archetype = strings.TrimSuffix(archetype, "/")
		archetype = c.Path + "archetypes/" + archetype

		// Check if the archetype ending with .markdown exists
		if _, err := os.Stat(archetype + ".markdown"); err == nil {
			err = utils.CopyFile(archetype+".markdown", filename)
			if err != nil {
				w.Write([]byte(err.Error()))
				return http.StatusInternalServerError, err
			}

			w.Header().Set("Location", "/admin/edit/"+filename)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{}"))
			return 201, nil
		}

		// Check if the archetype ending with .md exists
		if _, err := os.Stat(archetype + ".md"); err == nil {
			err = utils.CopyFile(archetype+".md", filename)
			if err != nil {
				w.Write([]byte(err.Error()))
				return http.StatusInternalServerError, err
			}

			w.Header().Set("Location", "/admin/edit/"+filename)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{}"))
			return 201, nil
		}
	}

	wf, err := os.Create(filename)
	if err != nil {
		w.Write([]byte(err.Error()))
		return http.StatusInternalServerError, err
	}

	defer wf.Close()

	w.Header().Set("Location", "/admin/edit/"+filename)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
	return http.StatusOK, nil
}

func upload(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Parse the multipart form in the request
	err := r.ParseMultipartForm(100000)
	if err != nil {
		w.Write([]byte(err.Error()))
		return http.StatusInternalServerError, err
	}

	// For each file header in the multipart form
	for _, fheaders := range r.MultipartForm.File {
		// Handle each file
		for _, hdr := range fheaders {
			// Open the first file
			var infile multipart.File
			if infile, err = hdr.Open(); nil != err {
				w.Write([]byte(err.Error()))
				return http.StatusInternalServerError, err
			}

			// Create the file
			var outfile *os.File
			if outfile, err = os.Create(c.Path + r.URL.Path + hdr.Filename); nil != err {
				w.Write([]byte(err.Error()))
				return http.StatusInternalServerError, err
			}

			// Copy the file content
			if _, err = io.Copy(outfile, infile); nil != err {
				w.Write([]byte(err.Error()))
				return http.StatusInternalServerError, err
			}

			defer outfile.Close()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
	return http.StatusOK, nil
}

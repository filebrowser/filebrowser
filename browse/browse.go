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
	"text/template"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/utils"
	"github.com/mholt/caddy/middleware"
	"github.com/mholt/caddy/middleware/browse"
)

// ServeHTTP is used to serve the content of Browse page
// using Browse middleware from Caddy
func ServeHTTP(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Removes the page main path from the URL
	r.URL.Path = strings.Replace(r.URL.Path, "/admin/browse", "", 1)

	switch r.Method {
	case "DELETE":
		return delete(w, r)
	case "POST":
		return post(w, r)
	case "GET":
		return get(w, r, c)
	default:
		return 400, nil
	}
}

func delete(w http.ResponseWriter, r *http.Request) (int, error) {
	// Remove both beginning and trailing slashes
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/")
	r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")

	// Check if the file or directory exists
	if stat, err := os.Stat(r.URL.Path); err == nil {
		var err error
		// If it's dir, remove all of the content inside
		if stat.IsDir() {
			err = os.RemoveAll(r.URL.Path)
		} else {
			err = os.Remove(r.URL.Path)
		}

		// Check for errors
		if err != nil {
			w.Write([]byte(err.Error()))
			return 500, err
		}
	} else {
		return 404, nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
	return 200, nil
}

func post(w http.ResponseWriter, r *http.Request) (int, error) {
	// Remove both beginning  slashes
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/")

	// If it's the upload of a file
	if r.Header.Get("X-Upload") == "true" {
		return upload(w, r)
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
		return 400, errors.New("Filename not specified.")
	}

	if _, ok := info["archetype"]; !ok {
		return 400, errors.New("Archtype not specified.")
	}

	// Sanitize the file name path
	filename := info["filename"].(string)
	filename = strings.TrimPrefix(filename, "/")
	filename = strings.TrimSuffix(filename, "/")
	filename = r.URL.Path + filename

	// Check if the archetype is defined
	if info["archetype"] != "" {
		// Sanitize the archetype path
		archetype := info["archetype"].(string)
		archetype = strings.Replace(archetype, "/archetypes", "", 1)
		archetype = strings.Replace(archetype, "archetypes", "", 1)
		archetype = strings.TrimPrefix(archetype, "/")
		archetype = strings.TrimSuffix(archetype, "/")
		archetype = "archetypes/" + archetype

		// Check if the archetype ending with .markdown exists
		if _, err := os.Stat(archetype + ".markdown"); err == nil {
			err = utils.CopyFile(archetype+".markdown", filename)
			if err != nil {
				w.Write([]byte(err.Error()))
				return 500, err
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
				return 500, err
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
		return 500, err
	}

	defer wf.Close()

	w.Header().Set("Location", "/admin/edit/"+filename)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
	return 200, nil
}

func upload(w http.ResponseWriter, r *http.Request) (int, error) {
	// Parse the multipart form in the request
	err := r.ParseMultipartForm(100000)
	if err != nil {
		w.Write([]byte(err.Error()))
		return 500, err
	}

	// For each file header in the multipart form
	for _, fheaders := range r.MultipartForm.File {
		// Handle each file
		for _, hdr := range fheaders {
			// Open the first file
			var infile multipart.File
			if infile, err = hdr.Open(); nil != err {
				w.Write([]byte(err.Error()))
				return 500, err
			}

			// Create the file
			var outfile *os.File
			if outfile, err = os.Create(r.URL.Path + hdr.Filename); nil != err {
				w.Write([]byte(err.Error()))
				return 500, err
			}

			// Copy the file content
			if _, err = io.Copy(outfile, infile); nil != err {
				w.Write([]byte(err.Error()))
				return 500, err
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
	return 200, nil
}

func get(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	functions := template.FuncMap{
		"CanBeEdited": utils.CanBeEdited,
		"Defined":     utils.Defined,
	}

	tpl, err := utils.GetTemplate(r, functions, "browse")

	if err != nil {
		w.Write([]byte(err.Error()))
		return 500, err
	}

	b := browse.Browse{
		Next: middleware.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 404, nil
		}),
		Root: "./",
		Configs: []browse.Config{
			{
				PathScope: "/",
				Variables: c,
				Template:  tpl,
			},
		},
		IgnoreIndexes: true,
	}

	return b.ServeHTTP(w, r)
}

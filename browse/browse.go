package browse

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
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

	// If the URL is blank now, replace it with a trailing slash
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	if r.Method == "DELETE" {
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
				return 500, err
			}
		} else {
			return 404, nil
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	} else if r.Method == "POST" {
		// Get the JSON information sent using a buffer
		buffer := new(bytes.Buffer)
		buffer.ReadFrom(r.Body)

		// Creates the raw file "map" using the JSON
		var info map[string]interface{}
		json.Unmarshal(buffer.Bytes(), &info)

		// Check if filename and archtype are specified in
		// the request
		if _, ok := info["filename"]; !ok {
			return 400, errors.New("Filename not specified.")
		}

		if _, ok := info["archtype"]; !ok {
			return 400, errors.New("Archtype not specified.")
		}

	} else {
		functions := template.FuncMap{
			"CanBeEdited": utils.CanBeEdited,
			"Defined":     utils.Defined,
		}

		tpl, err := utils.GetTemplate(r, functions, "browse")

		if err != nil {
			log.Print(err)
			return 500, err
		}

		b := browse.Browse{
			Next: middleware.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
				return 404, nil
			}),
			Root: "./",
			Configs: []browse.Config{
				browse.Config{
					PathScope: "/",
					Variables: c,
					Template:  tpl,
				},
			},
			IgnoreIndexes: true,
		}

		return b.ServeHTTP(w, r)
	}

	return 200, nil
}

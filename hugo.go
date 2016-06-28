//go:generate go get github.com/jteeuwen/go-bindata
//go:generate go install github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -debug -pkg hugo -prefix "assets" -o binary.go assets/...

// Package hugo makes the bridge between the static website generator Hugo
// and the webserver Caddy, also providing an administrative user interface.
package hugo

import (
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager"
	"github.com/hacdias/caddy-filemanager/assets"
	"github.com/hacdias/caddy-filemanager/directory"
	"github.com/hacdias/caddy-filemanager/utils/variables"
	"github.com/hacdias/caddy-hugo/utils/commands"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// Hugo is hugo
type Hugo struct {
	Next        httpserver.Handler
	Config      *Config
	FileManager *filemanager.FileManager
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
func (h Hugo) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	// If the site matches the baseURL
	if httpserver.Path(r.URL.Path).Matches(h.Config.BaseURL) {
		// Serve the hugo assets
		if httpserver.Path(r.URL.Path).Matches(h.Config.BaseURL + AssetsURL) {
			return serveAssets(w, r, h.Config)
		}

		// Serve the filemanager assets
		if httpserver.Path(r.URL.Path).Matches(h.Config.BaseURL + assets.BaseURL) {
			return h.FileManager.ServeHTTP(w, r)
		}

		// If the url matches exactly with /{admin}/settings/ serve that page
		// page variable isn't used here to avoid people using URLs like
		// "/{admin}/settings/something".
		if r.URL.Path == h.Config.BaseURL+"/settings/" || r.URL.Path == h.Config.BaseURL+"/settings" {
			var frontmatter string
			var err error

			if _, err = os.Stat(h.Config.Root + "config.yaml"); err == nil {
				frontmatter = "yaml"
			}

			if _, err = os.Stat(h.Config.Root + "config.json"); err == nil {
				frontmatter = "json"
			}

			if _, err = os.Stat(h.Config.Root + "config.toml"); err == nil {
				frontmatter = "toml"
			}

			http.Redirect(w, r, h.Config.BaseURL+"/config."+frontmatter, http.StatusTemporaryRedirect)
			return 0, nil
		}

		if r.Method == http.MethodPost && r.Header.Get("archetype") != "" {

			return 0, nil
		}

		if directory.CanBeEdited(r.URL.Path) && r.Method == http.MethodPut {
			code, err := h.FileManager.ServeHTTP(w, r)

			if err != nil {
				return code, err
			}

			if r.Header.Get("Regenerate") == "true" {
				RunHugo(h.Config, false)
			}

			if r.Header.Get("Schedule") != "" {

			}

			return code, err
		}

		return h.FileManager.ServeHTTP(w, r)
	}

	return h.Next.ServeHTTP(w, r)
}

// RunHugo is used to run the static website generator
func RunHugo(c *Config, force bool) {
	os.RemoveAll(c.Root + "public")

	// Prevent running if watching is enabled
	if b, pos := variables.StringInSlice("--watch", c.Args); b && !force {
		if len(c.Args) > pos && c.Args[pos+1] != "false" {
			return
		}

		if len(c.Args) == pos+1 {
			return
		}
	}

	if err := commands.Run(c.Hugo, c.Args, c.Root); err != nil {
		log.Panic(err)
	}
}

// serveAssets provides the needed assets for the front-end
func serveAssets(w http.ResponseWriter, r *http.Request, c *Config) (int, error) {
	// gets the filename to be used with Assets function
	filename := strings.Replace(r.URL.Path, c.BaseURL+AssetsURL, "public", 1)
	file, err := Asset(filename)
	if err != nil {
		return http.StatusNotFound, nil
	}

	// Get the file extension and its mimetype
	extension := filepath.Ext(filename)
	mediatype := mime.TypeByExtension(extension)

	// Write the header with the Content-Type and write the file
	// content to the buffer
	w.Header().Set("Content-Type", mediatype)
	w.Write(file)
	return 200, nil
}

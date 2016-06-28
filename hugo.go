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

func (h Hugo) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if httpserver.Path(r.URL.Path).Matches(h.Config.BaseURL) {
		if httpserver.Path(r.URL.Path).Matches(h.Config.BaseURL + AssetsURL) {

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

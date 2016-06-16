//go:generate go get github.com/jteeuwen/go-bindata
//go:generate go install github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -prefix assets/ -pkg assets -o routes/assets/assets.go assets/templates/ assets/public/...

// Package hugo makes the bridge between the static website generator Hugo
// and the webserver Caddy, also providing an administrative user interface.
package hugo

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/routes/assets"
	"github.com/hacdias/caddy-hugo/routes/browse"
	"github.com/hacdias/caddy-hugo/routes/editor"
	"github.com/hacdias/caddy-hugo/routes/errors"
	"github.com/hacdias/caddy-hugo/routes/git"
	"github.com/hacdias/caddy-hugo/tools/server"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// Hugo contais the next middleware to be run and the configuration
// of the current one.
type Hugo struct {
	Next   httpserver.Handler
	Config *config.Config
}

// ServeHTTP is the main function of the whole plugin that routes every single
// request to its function.
func (h Hugo) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	// Only handle /{admin} path
	if httpserver.Path(r.URL.Path).Matches(h.Config.Admin) {
		var err error
		var page string
		code := 404

		// If the length of the components string is less than one, the variable
		// page will always be "admin"
		if len(server.ParseURLComponents(r)) > 1 {
			page = server.ParseURLComponents(r)[1]
		} else {
			page = server.ParseURLComponents(r)[0]
		}

		// If the page isn't "assets" neither "edit", it should always put a
		// trailing slash in the path
		if page != "assets" && page != "edit" {
			if r.URL.Path[len(r.URL.Path)-1] != '/' {
				http.Redirect(w, r, r.URL.Path+"/", http.StatusTemporaryRedirect)
				return 0, nil
			}
		}

		// If the current page is only "/{admin}/", redirect to "/{admin}/browse/content/"
		if r.URL.Path == h.Config.Admin+"/" {
			http.Redirect(w, r, h.Config.Admin+"/browse/content/", http.StatusTemporaryRedirect)
			return 0, nil
		}

		// If the url matches exactly with /{admin}/settings/ serve that page
		// page variable isn't used here to avoid people using URLs like
		// "/{admin}/settings/something".
		if r.URL.Path == h.Config.Admin+"/settings/" {
			var frontmatter string

			if _, err = os.Stat(h.Config.Path + "config.yaml"); err == nil {
				frontmatter = "yaml"
			}

			if _, err = os.Stat(h.Config.Path + "config.json"); err == nil {
				frontmatter = "json"
			}

			if _, err = os.Stat(h.Config.Path + "config.toml"); err == nil {
				frontmatter = "toml"
			}

			http.Redirect(w, r, h.Config.Admin+"/edit/config."+frontmatter, http.StatusTemporaryRedirect)
			return 0, nil
		}

		// Serve the static assets
		if page == "assets" {
			code, err = serveAssets(w, r, h.Config)
		}

		// Browse page
		if page == "browse" {
			code, err = browse.ServeHTTP(w, r, h.Config)
		}

		// Edit page
		if page == "edit" {
			code, err = editor.ServeHTTP(w, r, h.Config)
		}

		// Git API
		if page == "git" {
			code, err = git.ServeHTTP(w, r, h.Config)
		}

		if code != 0 && code != 200 {
			code, err = errors.ServeHTTP(w, r, code, err)
		}

		return code, err
	}

	return h.Next.ServeHTTP(w, r)
}

// serveAssets handles the /{admin}/assets requests
func serveAssets(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	filename := strings.Replace(r.URL.Path, c.Admin+"/assets", "public", 1)
	file, err := assets.Asset(filename)

	if err != nil {
		return 404, nil
	}

	// Get the file extension ant its mime type
	extension := filepath.Ext(filename)
	mime := mime.TypeByExtension(extension)

	// Write the header with the Content-Type and write the file
	// content to the buffer
	w.Header().Set("Content-Type", mime)
	w.Write(file)
	return 200, nil
}

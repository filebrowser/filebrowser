//go:generate go-bindata -pkg assets -o assets/assets.go templates/ assets/css/ assets/js/ assets/fonts/

package hugo

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-hugo/assets"
	"github.com/hacdias/caddy-hugo/browse"
	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/editor"
	"github.com/hacdias/caddy-hugo/utils"
	"github.com/mholt/caddy/config/setup"
	"github.com/mholt/caddy/middleware"
)

// Setup configures the middleware
func Setup(c *setup.Controller) (middleware.Middleware, error) {
	config, _ := config.ParseHugo(c)
	utils.RunHugo(config)

	return func(next middleware.Handler) middleware.Handler {
		return &CaddyHugo{Next: next, Config: config}
	}, nil
}

// CaddyHugo main type
type CaddyHugo struct {
	Next   middleware.Handler
	Config *config.Config
}

func (h CaddyHugo) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	// Only handle /admin path
	if middleware.Path(r.URL.Path).Matches("/admin") {
		var err error
		var page string
		code := 404

		// If the length of the components string is less than one, the variable
		// page will always be "admin"
		if len(utils.ParseComponents(r)) > 1 {
			page = utils.ParseComponents(r)[1]
		} else {
			page = utils.ParseComponents(r)[0]
		}

		// If the page isn't "assets" neither "edit", it should always put a
		// trailing slash in the path
		if page != "assets" && page != "edit" {
			if r.URL.Path[len(r.URL.Path)-1] != '/' {
				http.Redirect(w, r, r.URL.Path+"/", http.StatusTemporaryRedirect)
				return 0, nil
			}
		}

		// If the current page is only "/admin/", redirect to "/admin/browse/contents"
		if r.URL.Path == "/admin/" {
			http.Redirect(w, r, "/admin/browse/content/", http.StatusTemporaryRedirect)
			return 0, nil
		}

		// Serve the static assets
		if page == "assets" {
			filename := strings.Replace(r.URL.Path, "/admin/", "", 1)
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

		// If the url matches exactly with /admin/settings/ serve that page
		// page variable isn't used here to avoid people using URLs like
		// "/admin/settings/something".
		if r.URL.Path == "/admin/settings/" {
			var frontmatter string

			if _, err := os.Stat("config.yaml"); err == nil {
				frontmatter = "yaml"
			}

			if _, err := os.Stat("config.json"); err == nil {
				frontmatter = "json"
			}

			if _, err := os.Stat("config.toml"); err == nil {
				frontmatter = "toml"
			}

			http.Redirect(w, r, "/admin/edit/config."+frontmatter, http.StatusTemporaryRedirect)
			return 0, nil
		}

		// Browse page
		if page == "browse" {
			code, err = browse.ServeHTTP(w, r, h.Config)
		}

		// Edit page
		if page == "edit" {
			code, err = editor.ServeHTTP(w, r, h.Config)
		}

		// Whenever the header "X-Refenerate" is true, the website should be
		// regenerated. Used in edit and settings, for example.
		if r.Header.Get("X-Regenerate") == "true" {
			utils.RunHugo(h.Config)
		}

		return code, err
	}

	return h.Next.ServeHTTP(w, r)
}

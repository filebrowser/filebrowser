//go:generate go get github.com/jteeuwen/go-bindata
//go:generate go install github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -debug -pkg assets -o assets/assets.go templates/ assets/css/ assets/js/ assets/fonts/

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

	"github.com/hacdias/caddy-hugo/assets"
	"github.com/hacdias/caddy-hugo/browse"
	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/editor"
	"github.com/hacdias/caddy-hugo/git"
	"github.com/hacdias/caddy-hugo/utils"
	"github.com/mholt/caddy/caddy/setup"
	"github.com/mholt/caddy/middleware"
)

// Setup is the init function of Caddy plugins and it configures the whole
// middleware thing.
func Setup(c *setup.Controller) (middleware.Middleware, error) {
	config, _ := config.ParseHugo(c)

	// Checks if there is an Hugo website in the path that is provided.
	// If not, a new website will be created.
	create := true

	if _, err := os.Stat(config.Path + "config.yaml"); err == nil {
		create = false
	}

	if _, err := os.Stat(config.Path + "config.json"); err == nil {
		create = false
	}

	if _, err := os.Stat(config.Path + "config.toml"); err == nil {
		create = false
	}

	if create {
		err := utils.RunCommand(config.Hugo, []string{"new", "site", config.Path, "--force"}, ".")
		if err != nil {
			log.Panic(err)
		}
	}

	// Generates the Hugo website for the first time the plugin is activated.
	go utils.Run(config, true)

	return func(next middleware.Handler) middleware.Handler {
		return &CaddyHugo{Next: next, Config: config}
	}, nil
}

// CaddyHugo contais the next middleware to be run and the configuration
// of the current one.
type CaddyHugo struct {
	Next   middleware.Handler
	Config *config.Config
}

// ServeHTTP is the main function of the whole plugin that routes every single
// request to its function.
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

		// If the current page is only "/admin/", redirect to "/admin/browse/content/"
		if r.URL.Path == "/admin/" {
			http.Redirect(w, r, "/admin/browse/content/", http.StatusTemporaryRedirect)
			return 0, nil
		}

		// If the url matches exactly with /admin/settings/ serve that page
		// page variable isn't used here to avoid people using URLs like
		// "/admin/settings/something".
		if r.URL.Path == "/admin/settings/" {
			var frontmatter string

			if _, err := os.Stat(h.Config.Path + "config.yaml"); err == nil {
				frontmatter = "yaml"
			}

			if _, err := os.Stat(h.Config.Path + "config.json"); err == nil {
				frontmatter = "json"
			}

			if _, err := os.Stat(h.Config.Path + "config.toml"); err == nil {
				frontmatter = "toml"
			}

			http.Redirect(w, r, "/admin/edit/config."+frontmatter, http.StatusTemporaryRedirect)
			return 0, nil
		}

		// Serve the static assets
		if page == "assets" {
			code, err = serveAssets(w, r)
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

		// Whenever the header "X-Regenerate" is true, the website should be
		// regenerated. Used in edit and settings, for example.
		if r.Header.Get("X-Regenerate") == "true" {
			go utils.Run(h.Config, false)
		}

		if err != nil {
			log.Panic(err)
		}

		/*

			// Create the functions map, then the template, check for erros and
			// execute the template if there aren't errors
			functions := template.FuncMap{
				"Defined": utils.Defined,
			}

			switch code {
			case 404:
				tpl, _ := utils.GetTemplate(r, functions, "404")
				tpl.Execute(w, nil)
				code = 200
				err = nil
			} */

		return code, err
	}

	return h.Next.ServeHTTP(w, r)
}

// serveAssets handles the /admin/assets requests
func serveAssets(w http.ResponseWriter, r *http.Request) (int, error) {
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

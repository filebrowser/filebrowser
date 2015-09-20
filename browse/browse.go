package browse

import (
	"log"
	"net/http"
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

package browse

import (
	"net/http"
	"text/template"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/tools/templates"
	"github.com/hacdias/caddy-hugo/tools/variables"
	"github.com/mholt/caddy/caddyhttp/browse"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// GET handles the GET method on browse page and shows the files listing Using
// the Browse Caddy middleware.
func GET(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	functions := template.FuncMap{
		"CanBeEdited": templates.CanBeEdited,
		"Defined":     variables.Defined,
	}

	tpl, err := templates.Get(r, functions, "browse")

	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Using Caddy's Browse middleware
	b := browse.Browse{
		Next: httpserver.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 404, nil
		}),
		Configs: []browse.Config{
			{
				PathScope: "/",
				Root:      http.Dir(c.Path),
				Variables: c,
				Template:  tpl,
			},
		},
		IgnoreIndexes: true,
	}

	return b.ServeHTTP(w, r)
}

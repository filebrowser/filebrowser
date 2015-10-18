package browse

import (
	"net/http"
	"text/template"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/utils"
	"github.com/mholt/caddy/middleware"
	"github.com/mholt/caddy/middleware/browse"
)

// GET handles the GET method on browse page
func GET(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	functions := template.FuncMap{
		"CanBeEdited": utils.CanBeEdited,
		"Defined":     utils.Defined,
	}

	tpl, err := utils.GetTemplate(r, functions, "browse")

	if err != nil {
		return http.StatusInternalServerError, err
	}

	b := browse.Browse{
		Next: middleware.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 404, nil
		}),
		Root: c.Path,
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

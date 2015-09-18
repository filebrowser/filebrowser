package browse

import (
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/hacdias/caddy-hugo/editor"
	"github.com/hacdias/caddy-hugo/utils"
	"github.com/mholt/caddy/middleware"
	"github.com/mholt/caddy/middleware/browse"
)

// Execute sth
func Execute(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path[len(r.URL.Path)-1] != '/' {
		http.Redirect(w, r, r.URL.Path+"/", http.StatusTemporaryRedirect)
		return 0, nil
	}

	r.URL.Path = strings.Replace(r.URL.Path, "/admin/browse", "", 1)

	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	functions := template.FuncMap{
		"canBeEdited": editor.CanBeEdited,
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
				Template:  tpl,
			},
		},
		IgnoreIndexes: true,
	}

	return b.ServeHTTP(w, r)
}

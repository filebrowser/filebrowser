package browse

import (
	"log"
	"net/http"
	"strings"

	"github.com/hacdias/caddy-hugo/page"
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

	tpl, err := page.GetTemplate(r, "browse")

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
	}

	return b.ServeHTTP(w, r)
}

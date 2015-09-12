package hugo

import (
	"net/http"
	"strings"

	"github.com/spf13/hugo/commands"

	"github.com/mholt/caddy/config/setup"
	"github.com/mholt/caddy/middleware"
)

// Setup function
func Setup(c *setup.Controller) (middleware.Middleware, error) {
	commands.Execute()

	return func(next middleware.Handler) middleware.Handler {
		return &handler{Next: next}
	}, nil
}

type handler struct{ Next middleware.Handler }

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if middleware.Path(r.URL.Path).Matches("/admin") {

		if middleware.Path(r.URL.Path).Matches("/admin/new") {
			w.Write([]byte("New"))
		} else if middleware.Path(r.URL.Path).Matches("/admin/edit") {
			var fileName string

			fileName = strings.Replace(r.URL.Path, "/admin/edit", "", 1)

			w.Write([]byte("Edit " + fileName))
		} else {
			w.Write([]byte("Admin"))
		}

		return 200, nil
	}

	return h.Next.ServeHTTP(w, r)
}

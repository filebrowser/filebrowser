package hugo

import (
	"net/http"

	"github.com/spf13/hugo/commands"

	"github.com/mholt/caddy/config/setup"
	"github.com/mholt/caddy/middleware"
)

func Setup(c *setup.Controller) (middleware.Middleware, error) {
	for c.Next() {
		commands.Execute()
	}

	return func(next middleware.Handler) middleware.Handler {
		return &handler{}
	}, nil
}

type handler struct{}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	http.ServeFile(w, r, "public"+r.URL.Path)
	return 200, nil
}

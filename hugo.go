package hugo

import (
	"net/http"

	"github.com/hacdias/caddy-hugo/routing"
	"github.com/mholt/caddy/config/setup"
	"github.com/mholt/caddy/middleware"
	"github.com/spf13/hugo/commands"
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
		return routing.Route(w, r)
	}

	return h.Next.ServeHTTP(w, r)
}

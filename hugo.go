package hugo

import (
	"net/http"

	"github.com/spf13/hugo/commands"

	"github.com/mholt/caddy/config/setup"
	"github.com/mholt/caddy/middleware"
)

// Setup function
func Setup(c *setup.Controller) (middleware.Middleware, error) {
	commands.Execute()

	return func(next middleware.Handler) middleware.Handler {
		return &handler{
			Next: next,
		}
	}, nil
}

type handler struct{ Next middleware.Handler }
type adminHandler struct{}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	// do path matching
	if middleware.Path(r.URL.Path).Matches("/admin") {
		a := new(adminHandler)
		return a.ServeHTTP(w, r)
	}

	return 200, nil
}

func (a adminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	w.Write([]byte("Admin area"))
	return 200, nil
}

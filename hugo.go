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
		return &handler{}
	}, nil
}

type handler struct{}
type adminHandler struct{}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {

	// do path matching
	if middleware.Path(r.URL.Path).Matches("/admin") {
		a := new(adminHandler)
		return a.ServeHTTP(w, r)
	}
	http.ServeFile(w, r, "public"+r.URL.Path)

	return 200, nil
}

func (a adminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	w.Write([]byte("Admin area"))
	return 200, nil
}

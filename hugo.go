package hugo

import (
	"net/http"

	"github.com/mholt/caddy/config/setup"
	"github.com/mholt/caddy/middleware"
)

func Setup(c *setup.Controller) (middleware.Middleware, error) {
	return func(next middleware.Handler) middleware.Handler {
		return &handler{}
	}, nil
}

type handler struct{}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	w.Write([]byte("Hello, I'm a caddy middleware"))
	return 200, nil
}

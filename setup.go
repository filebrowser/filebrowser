package filemanager

import (
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("filemanager", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

// setup configures the middlware.
func setup(c *caddy.Controller) error {
	cnf := httpserver.GetConfig(c.Key)

	// parse config

	mid := func(next httpserver.Handler) httpserver.Handler {
		return FileManager{
			Next: next,
		}
	}

	cnf.AddMiddleware(mid)
	return nil
}

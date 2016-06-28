package filemanager

import (
	"github.com/hacdias/caddy-filemanager/config"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("filemanager", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

// setup configures a new FileManager middleware instance.
func setup(c *caddy.Controller) error {
	configs, err := config.Parse(c)
	if err != nil {
		return err
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return FileManager{Configs: configs, Next: next}
	})

	return nil
}

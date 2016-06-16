package hugo

import (
	"log"
	"os"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/tools/commands"
	"github.com/hacdias/caddy-hugo/tools/hugo"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("hugo", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

// Setup is the init function of Caddy plugins and it configures the whole
// middleware thing.
func setup(c *caddy.Controller) error {
	cnf := httpserver.GetConfig(c.Key)
	conf, _ := config.Parse(c, cnf.Root)

	// Checks if there is an Hugo website in the path that is provided.
	// If not, a new website will be created.
	create := true

	if _, err := os.Stat(conf.Path + "config.yaml"); err == nil {
		create = false
	}

	if _, err := os.Stat(conf.Path + "config.json"); err == nil {
		create = false
	}

	if _, err := os.Stat(conf.Path + "config.toml"); err == nil {
		create = false
	}

	if create {
		err := commands.Run(conf.Hugo, []string{"new", "site", conf.Path, "--force"}, ".")
		if err != nil {
			log.Panic(err)
		}
	}

	// Generates the Hugo website for the first time the plugin is activated.
	go hugo.Run(conf, true)

	mid := func(next httpserver.Handler) httpserver.Handler {
		return &Hugo{Next: next, Config: conf}
	}

	cnf.AddMiddleware(mid)
	return nil
}

package hugo

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager"
	"github.com/hacdias/caddy-hugo/tools/commands"
	"github.com/hacdias/caddy-hugo/tools/installer"
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
	conf, _ := ParseHugo(c, cnf.Root)

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
	go Run(conf, true)

	mid := func(next httpserver.Handler) httpserver.Handler {
		return &Hugo{
			Next: next,
			FileManager: &filemanager.FileManager{
				Next: next,
				Configs: []filemanager.Config{
					filemanager.Config{
						PathScope: conf.Path,
						Root:      http.Dir(conf.Path),
						BaseURL:   conf.Admin,
					},
				},
			},
			Config: conf,
		}
	}

	cnf.AddMiddleware(mid)
	return nil
}

// Config is the add-on configuration set on Caddyfile
type Config struct {
	Public string   // Public content path
	Path   string   // Hugo files path
	Styles string   // Admin styles path
	Args   []string // Hugo arguments
	Hugo   string   // Hugo executable path
	Admin  string   // Hugo admin URL
	Git    bool     // Is this site a git repository
}

// ParseHugo parses the configuration file
func ParseHugo(c *caddy.Controller, root string) (*Config, error) {
	conf := &Config{
		Public: strings.Replace(root, "./", "", -1),
		Admin:  "/admin",
		Path:   "./",
		Git:    false,
	}

	conf.Hugo = installer.GetPath()

	for c.Next() {
		args := c.RemainingArgs()

		switch len(args) {
		case 1:
			conf.Path = args[0]
			conf.Path = strings.TrimSuffix(conf.Path, "/")
			conf.Path += "/"
		}

		for c.NextBlock() {
			switch c.Val() {
			case "styles":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				conf.Styles = c.Val()
				// Remove the beginning slash if it exists or not
				conf.Styles = strings.TrimPrefix(conf.Styles, "/")
				// Add a beginning slash to make a
				conf.Styles = "/" + conf.Styles
			case "admin":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				conf.Admin = c.Val()
				conf.Admin = strings.TrimPrefix(conf.Admin, "/")
				conf.Admin = "/" + conf.Admin
			default:
				key := "--" + c.Val()
				value := "true"

				if c.NextArg() {
					value = c.Val()
				}

				conf.Args = append(conf.Args, key+"="+value)
			}
		}
	}

	if _, err := os.Stat(filepath.Join(conf.Path, ".git")); err == nil {
		conf.Git = true
	}

	return conf, nil
}

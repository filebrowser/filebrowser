package hugo

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager"
	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/directory"
	"github.com/hacdias/caddy-filemanager/frontmatter"
	"github.com/hacdias/caddy-hugo/installer"
	"github.com/hacdias/caddy-hugo/utils/commands"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// AssetsURL is the base url for the assets
const AssetsURL = "/_hugointernal"

func init() {
	caddy.RegisterPlugin("hugo", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

// Setup is the init function of Caddy plugins and it configures the whole
// middleware thing.
func setup(c *caddy.Controller) error {
	cnf := httpserver.GetConfig(c)
	conf, _ := parse(c, cnf.Root)

	format := "toml"

	// Checks if there is an Hugo website in the path that is provided.
	// If not, a new website will be created.
	create := true

	if _, err := os.Stat(conf.Root + "config.yaml"); err == nil {
		format = "yaml"
		create = false
	}

	if _, err := os.Stat(conf.Root + "config.json"); err == nil {
		format = "json"
		create = false
	}

	if _, err := os.Stat(conf.Root + "config.toml"); err == nil {
		format = "toml"
		create = false
	}

	if create {
		err := commands.Run(conf.Hugo, []string{"new", "site", conf.Root, "--force"}, ".")
		if err != nil {
			log.Panic(err)
		}
	}

	// Get Default FrontMatter
	bytes, err := ioutil.ReadFile(filepath.Clean(conf.Root + "/config." + format))

	if err != nil {
		log.Panic(err)
	}

	bytes = directory.AppendFrontMatterRune(bytes, format)
	f, err := frontmatter.Unmarshal(bytes)

	if err != nil {
		log.Panic(err)
	}

	if val, ok := f.(map[string]interface{})["metaDataFormat"]; ok {
		format = val.(string)
	}

	// Generates the Hugo website for the first time the plugin is activated.
	go RunHugo(conf, true)

	mid := func(next httpserver.Handler) httpserver.Handler {
		return &Hugo{
			Next:   next,
			Config: conf,
			FileManager: &filemanager.FileManager{
				Next: next,
				Configs: []config.Config{
					config.Config{
						HugoEnabled: true,
						PathScope:   conf.Root,
						FrontMatter: format,
						Root:        http.Dir(conf.Root),
						BaseURL:     conf.BaseURL,
						StyleSheet:  conf.Styles,
					},
				},
			},
		}
	}

	cnf.AddMiddleware(mid)
	return nil
}

// Config is a configuration for managing a particular hugo website.
type Config struct {
	Public      string   // Public content path
	Root        string   // Hugo files path
	Hugo        string   // Hugo executable location
	Styles      string   // Admin styles path
	Args        []string // Hugo arguments
	BaseURL     string   // BaseURL of admin interface
	FileManager *filemanager.FileManager
}

// Parse parses the configuration set by the user so it can be
// used by the middleware
func parse(c *caddy.Controller, root string) (*Config, error) {
	conf := &Config{
		Public:  strings.Replace(root, "./", "", -1),
		BaseURL: "/admin",
		Root:    "./",
	}

	conf.Hugo = installer.GetPath()
	for c.Next() {
		args := c.RemainingArgs()

		switch len(args) {
		case 1:
			conf.Root = args[0]
			conf.Root = strings.TrimSuffix(conf.Root, "/")
			conf.Root += "/"
		}

		for c.NextBlock() {
			switch c.Val() {
			case "styles":
				if !c.NextArg() {
					return conf, c.ArgErr()
				}
				tplBytes, err := ioutil.ReadFile(c.Val())
				if err != nil {
					return conf, err
				}
				conf.Styles = string(tplBytes)
			case "admin":
				if !c.NextArg() {
					return conf, c.ArgErr()
				}
				conf.BaseURL = c.Val()
				conf.BaseURL = strings.TrimPrefix(conf.BaseURL, "/")
				conf.BaseURL = "/" + conf.BaseURL
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

	return conf, nil
}

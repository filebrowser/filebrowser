package hugo

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager"
	"github.com/hacdias/caddy-hugo/installer"
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

	if _, err := os.Stat(conf.Root + "config.yaml"); err == nil {
		create = false
	}

	if _, err := os.Stat(conf.Root + "config.json"); err == nil {
		create = false
	}

	if _, err := os.Stat(conf.Root + "config.toml"); err == nil {
		create = false
	}

	if create {
		err := Run(conf.Hugo, []string{"new", "site", conf.Root, "--force"}, ".")
		if err != nil {
			log.Panic(err)
		}
	}

	// Generates the Hugo website for the first time the plugin is activated.
	go RunHugo(conf, true)

	mid := func(next httpserver.Handler) httpserver.Handler {
		return &Hugo{
			Next:   next,
			Config: conf,
			FileManager: &filemanager.FileManager{
				Next: next,
				Configs: []filemanager.Config{
					filemanager.Config{
						PathScope:  conf.Root,
						Root:       http.Dir(conf.Root),
						BaseURL:    conf.BaseURL,
						StyleSheet: conf.Styles,
					},
				},
			},
		}
	}

	cnf.AddMiddleware(mid)
	return nil
}

// Config contains the configuration of hugo plugin
type Config struct {
	Args    []string // Hugo arguments
	Git     bool     // Is this site a git repository
	BaseURL string   // Admin URL to listen on
	Hugo    string   // Hugo executable path
	Root    string   // Hugo website path
	Public  string   // Public content path
	Styles  string   // Admin stylesheet
}

// ParseHugo parses the configuration file
func ParseHugo(c *caddy.Controller, root string) (*Config, error) {
	conf := &Config{
		Public:  strings.Replace(root, "./", "", -1),
		BaseURL: "/admin",
		Root:    "./",
		Git:     false,
		Hugo:    installer.GetPath(),
	}

	stlsbytes, err := Asset("public/css/styles.css")

	if err != nil {
		return conf, err
	}

	conf.Styles = string(stlsbytes)

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
				stylesheet, err := ioutil.ReadFile(c.Val())
				if err != nil {
					return conf, err
				}
				conf.Styles += string(stylesheet)
			case "admin":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				conf.BaseURL = c.Val()
				// Remove the beginning slash if it exists or not
				conf.BaseURL = strings.TrimPrefix(conf.BaseURL, "/")
				// Add a beginning slash to make a
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

	if _, err := os.Stat(filepath.Join(conf.Root, ".git")); err == nil {
		conf.Git = true
	}

	return conf, nil
}

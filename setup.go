package filemanager

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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
	configs, err := parseConfiguration(c)
	if err != nil {
		return err
	}

	httpserver.GetConfig(c.Key).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return FileManager{Configs: configs, Next: next}
	})

	return nil
}

// Config is a configuration for browsing in a particualr path.
type Config struct {
	PathScope   string
	Root        http.FileSystem
	BaseURL     string
	StyleSheet  string // Costum stylesheet
	HugoEnabled bool   // This must be only used by Hugo plugin
}

// parseConfiguration parses the configuration set by the user so it can
// be used by the middleware
func parseConfiguration(c *caddy.Controller) ([]Config, error) {
	var configs []Config

	appendConfig := func(cfg Config) error {
		for _, c := range configs {
			if c.PathScope == cfg.PathScope {
				return fmt.Errorf("duplicate file managing config for %s", c.PathScope)
			}
		}
		configs = append(configs, cfg)
		return nil
	}

	for c.Next() {
		var cfg = Config{PathScope: ".", BaseURL: "", HugoEnabled: false}
		for c.NextBlock() {
			switch c.Val() {
			case "show":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				cfg.PathScope = c.Val()
				cfg.PathScope = strings.TrimSuffix(cfg.PathScope, "/")
			case "on":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				cfg.BaseURL = c.Val()
				cfg.BaseURL = strings.TrimPrefix(cfg.BaseURL, "/")
				cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")
				cfg.BaseURL = "/" + cfg.BaseURL
			case "styles":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				tplBytes, err := ioutil.ReadFile(c.Val())
				if err != nil {
					return configs, err
				}
				cfg.StyleSheet = string(tplBytes)
			}
		}

		cfg.Root = http.Dir(cfg.PathScope)
		if err := appendConfig(cfg); err != nil {
			return configs, err
		}
	}

	return configs, nil
}

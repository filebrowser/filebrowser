package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-hugo/hugo"
	"github.com/mholt/caddy/caddy/setup"
)

// Config is the add-on configuration set on Caddyfile
type Config struct {
	Public string   // Public content path
	Path   string   // Hugo files path
	Styles string   // Admin styles path
	Args   []string // Hugo arguments
	Hugo   string   // Hugo executable path
	Git    bool     // Is this site a git repository
}

// ParseHugo parses the configuration file
func ParseHugo(c *setup.Controller) (*Config, error) {
	conf := &Config{
		Public: strings.Replace(c.Root, "./", "", -1),
		Path:   "./",
		Git:    false,
	}

	conf.Hugo = hugo.GetPath()

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

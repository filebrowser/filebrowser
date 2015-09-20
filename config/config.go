package config

import (
	"strings"

	"github.com/mholt/caddy/config/setup"
)

// Config is the add-on configuration set on Caddyfile
type Config struct {
	Styles string
}

// ParseHugo parses the configuration file
func ParseHugo(c *setup.Controller) (*Config, error) {
	conf := &Config{
		Styles: "",
	}

	for c.Next() {
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
			}
		}
	}

	return conf, nil
}

package config

import (
	"strings"

	"github.com/mholt/caddy/config/setup"
)

// Config is the add-on configuration set on Caddyfile
type Config struct {
	Styles string
	Flags  []string
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
			case "flags":
				conf.Flags = c.RemainingArgs()
				if len(conf.Flags) == 0 {
					return conf, c.ArgErr()
				}
			}
		}
	}

	conf.parseFlags()
	return conf, nil
}

func (c *Config) parseFlags() {
	for index, element := range c.Flags {
		c.Flags[index] = strings.Replace(element, "\"", "", -1)
	}
}

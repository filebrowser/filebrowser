package config

import (
	"strconv"
	"strings"

	"github.com/mholt/caddy/config/setup"
)

// Config is the add-on configuration set on Caddyfile
type Config struct {
	Styles  string
	Hugo    bool
	Args    []string
	Command string
}

// ParseCMS parses the configuration file
func ParseCMS(c *setup.Controller) (*Config, error) {
	conf := &Config{Hugo: true}

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
			case "hugo":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				var err error
				conf.Hugo, err = strconv.ParseBool(c.Val())
				if err != nil {
					return conf, err
				}
			case "command":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				conf.Command = c.Val()
			case "args":
				conf.Args = c.RemainingArgs()
				if len(conf.Args) == 0 {
					return conf, c.ArgErr()
				}
			}
		}
	}

	conf.parseFlags()
	return conf, nil
}

func (c *Config) parseFlags() {
	for index, element := range c.Args {
		c.Args[index] = strings.Replace(element, "\"", "", -1)
	}
}

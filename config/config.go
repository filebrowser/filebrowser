package config

import (
	"strings"

	"github.com/mholt/caddy/config/setup"
)

// Config is the add-on configuration set on Caddyfile
type Config struct {
	Styles  string
	Args    []string
	Command string
	Content string
}

// ParseCMS parses the configuration file
func ParseCMS(c *setup.Controller) (*Config, error) {
	conf := &Config{Content: "content"}

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
			case "content":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				conf.Content = c.Val()
				conf.Content = strings.TrimPrefix(conf.Content, "/")
				conf.Content = strings.TrimSuffix(conf.Content, "/")
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

	conf.parseArgs()
	return conf, nil
}

func (c *Config) parseArgs() {
	for index, element := range c.Args {
		c.Args[index] = strings.Replace(element, "\"", "", -1)
	}
}

package config

import (
	"strings"

	"github.com/mholt/caddy/config/setup"
)

// Config is the add-on configuration set on Caddyfile
type Config struct {
	Public  string
	Content string
	Path    string
	Styles  string
	Command string
	Hugo    bool
}

// ParseCMS parses the configuration file
func ParseCMS(c *setup.Controller) (*Config, error) {
	conf := &Config{
		Public:  strings.Replace(c.Root, "./", "", -1),
		Content: "content",
		Hugo:    true,
		Path:    "./",
	}

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
			case "content":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				conf.Content = c.Val()
			case "command":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				conf.Command = c.Val()

				if conf.Command != "" && !strings.HasPrefix(conf.Command, "-") {
					conf.Hugo = false
				}
			}
		}
	}

	return conf, nil
}

package config

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"runtime"
	"strings"

	"github.com/hacdias/caddy-hugo/insthugo"
	"github.com/mholt/caddy/caddy/setup"
)

// Config is the add-on configuration set on Caddyfile
type Config struct {
	Public string   // Public content path
	Path   string   // Hugo files path
	Styles string   // Admin styles path
	Args   []string // Hugo arguments
	Hugo   string   // Hugo executable path
}

// ParseHugo parses the configuration file
func ParseHugo(c *setup.Controller) (*Config, error) {
	// First check if Hugo is installed
	user, err := user.Current()

	if err != nil {
		log.Fatal(err)
	}

	conf := &Config{
		Public: strings.Replace(c.Root, "./", "", -1),
		Path:   "./",
		Hugo:   user.HomeDir + "/.caddy/bin/hugo",
	}

	if runtime.GOOS == "windows" {
		conf.Hugo += ".exe"
	}

	if _, err := os.Stat(conf.Hugo); os.IsNotExist(err) {
		fmt.Print("hey")
		insthugo.Install()
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

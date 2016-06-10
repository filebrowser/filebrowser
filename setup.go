package filemanager

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("filemanager", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

// setup configures a new Browse middleware instance.
func setup(c *caddy.Controller) error {
	// Second argument would be the template file to use
	tplBytes, err := Asset("templates/template.tmpl")
	if err != nil {
		return err
	}
	tplText := string(tplBytes)

	// Build the template
	tpl, err := template.New("listing").Parse(tplText)
	if err != nil {
		return err
	}
	Template = tpl

	configs, err := fileManagerParse(c)
	if err != nil {
		return err
	}

	f := FileManager{
		Configs:       configs,
		IgnoreIndexes: false,
	}

	httpserver.GetConfig(c.Key).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		f.Next = next
		return f
	})

	return nil
}

func fileManagerParse(c *caddy.Controller) ([]Config, error) {
	var configs []Config

	appendCfg := func(fmc Config) error {
		for _, c := range configs {
			if c.PathScope == fmc.PathScope {
				return fmt.Errorf("duplicate file managing config for %s", c.PathScope)
			}
		}
		configs = append(configs, fmc)
		return nil
	}

	for c.Next() {
		var fmc = Config{
			PathScope:  ".",
			BaseURL:    "/",
			StyleSheet: "",
		}

		for c.NextBlock() {
			switch c.Val() {
			case "show":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				fmc.PathScope = c.Val()
			case "on":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				fmc.BaseURL = c.Val()
			case "styles":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				tplBytes, err := ioutil.ReadFile(c.Val())
				if err != nil {
					return configs, err
				}
				fmc.StyleSheet = string(tplBytes)
			}
		}

		fmc.Root = http.Dir(fmc.PathScope)

		// Save configuration
		err := appendCfg(fmc)
		if err != nil {
			return configs, err
		}
	}

	return configs, nil
}

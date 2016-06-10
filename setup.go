package filemanager

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"

	"github.com/hacdias/caddy-filemanager/assets"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

const assetsURL = "/_filemanager_internal/"

func init() {
	caddy.RegisterPlugin("filemanager", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

// setup configures a new Browse middleware instance.
func setup(c *caddy.Controller) error {
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

	cfg := httpserver.GetConfig(c.Key)

	appendCfg := func(bc Config) error {
		for _, c := range configs {
			if c.PathScope == bc.PathScope {
				return fmt.Errorf("duplicate browsing config for %s", c.PathScope)
			}
		}
		configs = append(configs, bc)
		return nil
	}

	for c.Next() {
		var bc Config

		// First argument is directory to allow browsing; default is site root
		if c.NextArg() {
			bc.PathScope = c.Val()
		} else {
			bc.PathScope = "/"
		}
		bc.Root = http.Dir(cfg.Root)
		theRoot, err := bc.Root.Open("/") // catch a missing path early
		if err != nil {
			return configs, err
		}
		defer theRoot.Close()
		_, err = theRoot.Readdir(-1)
		if err != nil {
			return configs, err
		}

		var tplBytes []byte

		// Second argument would be the template file to use
		var tplText string
		if c.NextArg() {
			tplBytes, err = ioutil.ReadFile(c.Val())
			if err != nil {
				return configs, err
			}
			tplText = string(tplBytes)
		} else {
			tplBytes, err = assets.Asset(assetsURL + "template.tmpl")
			if err != nil {
				return configs, err
			}
			tplText = string(tplBytes)
		}

		// Build the template
		tpl, err := template.New("listing").Parse(tplText)
		if err != nil {
			return configs, err
		}
		bc.Template = tpl

		// Save configuration
		err = appendCfg(bc)
		if err != nil {
			return configs, err
		}
	}

	return configs, nil
}

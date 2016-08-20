package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// Config is a configuration for browsing in a particualr path.
type Config struct {
	*UserConfig
	BaseURL     string
	AbsoluteURL string
	AddrPath    string
	Token       string // Anti CSRF token
	FrontMatter string // Default frontmatter to save files in
	HugoEnabled bool   // Enables the Hugo plugin for File Manager
}

// UserConfig contains the configuration for each user
type UserConfig struct {
	PathScope   string
	Root        http.FileSystem
	StyleSheet  string // Costum stylesheet
	FrontMatter string // Default frontmatter to save files in
	AllowNew    bool   // Can create files and folders
	AllowEdit   bool   // Can edit/rename files

	Allow      []string         // Allowed browse directories/files
	AllowRegex []*regexp.Regexp // Regex of the previous
	Block      []string         // Blocked browse directories/files
	BlockRegex []*regexp.Regexp // Regex of the previous

	AllowCommands   bool     // Can execute commands
	AllowedCommands []string // Allowed commands
	BlockedCommands []string // Blocked Commands
}

// Parse parses the configuration set by the user so it can
// be used by the middleware
func Parse(c *caddy.Controller) ([]Config, error) {
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

	var err error

	for c.Next() {
		var cfg = Config{}
		cfg.PathScope = "."
		cfg.BaseURL = ""
		cfg.FrontMatter = "yaml"
		cfg.HugoEnabled = false

		for c.NextBlock() {
			switch c.Val() {
			case "show":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				cfg.PathScope = c.Val()
				cfg.PathScope = strings.TrimSuffix(cfg.PathScope, "/")
			case "frontmatter":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				cfg.FrontMatter = c.Val()
				if cfg.FrontMatter != "yaml" && cfg.FrontMatter != "json" && cfg.FrontMatter != "toml" {
					return configs, c.Err("frontmatter type not supported")
				}
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
				var tplBytes []byte
				tplBytes, err = ioutil.ReadFile(c.Val())
				if err != nil {
					return configs, err
				}
				cfg.StyleSheet = string(tplBytes)
			case "allow_new":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				cfg.AllowNew, err = strconv.ParseBool(c.Val())
				if err != nil {
					return configs, err
				}
			case "allow_edit":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				cfg.AllowEdit, err = strconv.ParseBool(c.Val())
				if err != nil {
					return configs, err
				}
			case "allow_comands":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				cfg.AllowCommands, err = strconv.ParseBool(c.Val())
				if err != nil {
					return configs, err
				}
			}
		}

		caddyConf := httpserver.GetConfig(c)
		cfg.AbsoluteURL = strings.TrimSuffix(caddyConf.Addr.Path, "/") + "/" + cfg.BaseURL
		cfg.AbsoluteURL = strings.Replace(cfg.AbsoluteURL, "//", "/", -1)
		cfg.AbsoluteURL = strings.TrimSuffix(cfg.AbsoluteURL, "/")
		cfg.AddrPath = strings.TrimSuffix(caddyConf.Addr.Path, "/")
		cfg.Root = http.Dir(cfg.PathScope)
		if err := appendConfig(cfg); err != nil {
			return configs, err
		}
	}

	return configs, nil
}

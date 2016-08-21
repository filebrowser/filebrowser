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
	HugoEnabled bool   // Enables the Hugo plugin for File Manager
	Users       map[string]*UserConfig
	CurrentUser *UserConfig
}

// UserConfig contains the configuration for each user
type UserConfig struct {
	PathScope     string          // Path the user have access
	Root          http.FileSystem // The virtual file system the user have access
	StyleSheet    string          // Costum stylesheet
	FrontMatter   string          // Default frontmatter to save files in
	AllowNew      bool            // Can create files and folders
	AllowEdit     bool            // Can edit/rename files
	AllowCommands bool            // Can execute commands
	Commands      []string        // Available Commands
	Rules         []*Rule         // Access rules
}

// Rule is a dissalow/allow rule
type Rule struct {
	Regex  bool
	Allow  bool
	Path   string
	Rexexp *regexp.Regexp
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
	var cCfg *UserConfig
	var baseURL string

	for c.Next() {
		var cfg = Config{UserConfig: &UserConfig{}}
		cfg.PathScope = "."
		cfg.BaseURL = ""
		cfg.FrontMatter = "yaml"
		cfg.HugoEnabled = false
		cfg.Users = map[string]*UserConfig{}
		cfg.AllowCommands = true
		cfg.AllowEdit = true
		cfg.AllowNew = true
		cfg.Commands = []string{"git", "svn", "hg"}

		baseURL = ""
		cCfg = cfg.UserConfig

		for c.NextBlock() {
			switch c.Val() {
			case "on":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				baseURL = c.Val()
			case "frontmatter":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				cCfg.FrontMatter = c.Val()
				if cCfg.FrontMatter != "yaml" && cCfg.FrontMatter != "json" && cCfg.FrontMatter != "toml" {
					return configs, c.Err("frontmatter type not supported")
				}
			case "show":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				cCfg.PathScope = c.Val()
				cCfg.PathScope = strings.TrimSuffix(cCfg.PathScope, "/")
			case "styles":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				var tplBytes []byte
				tplBytes, err = ioutil.ReadFile(c.Val())
				if err != nil {
					return configs, err
				}
				cCfg.StyleSheet = string(tplBytes)
			case "allow_new":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				cCfg.AllowNew, err = strconv.ParseBool(c.Val())
				if err != nil {
					return configs, err
				}
			case "allow_edit":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				cCfg.AllowEdit, err = strconv.ParseBool(c.Val())
				if err != nil {
					return configs, err
				}
			case "allow_commands":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				cCfg.AllowCommands, err = strconv.ParseBool(c.Val())
				if err != nil {
					return configs, err
				}
			case "allow_command":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				cCfg.Commands = append(cCfg.Commands, c.Val())
			case "block_command":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				index := 0

				for i, val := range cCfg.Commands {
					if val == c.Val() {
						index = i
					}
				}

				cCfg.Commands = append(cCfg.Commands[:index], cCfg.Commands[index+1:]...)
			case "allow":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				cCfg.Rules = append(cCfg.Rules, &Rule{
					Regex:  false,
					Allow:  true,
					Path:   c.Val(),
					Rexexp: nil,
				})
			case "allow_r":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				cCfg.Rules = append(cCfg.Rules, &Rule{
					Regex:  true,
					Allow:  true,
					Path:   "",
					Rexexp: regexp.MustCompile(c.Val()),
				})
			case "block":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				cCfg.Rules = append(cCfg.Rules, &Rule{
					Regex:  false,
					Allow:  false,
					Path:   c.Val(),
					Rexexp: nil,
				})
			case "block_r":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				cCfg.Rules = append(cCfg.Rules, &Rule{
					Regex:  true,
					Allow:  false,
					Path:   "",
					Rexexp: regexp.MustCompile(c.Val()),
				})
			// NEW USER BLOCK?
			default:
				cCfg.Root = http.Dir(cCfg.PathScope)

				val := c.Val()
				// Checks if it's a new user
				if !strings.HasSuffix(val, ":") {
					fmt.Println("Unknown option " + val)
				}

				// Get the username, sets the current user, and initializes it
				val = strings.TrimSuffix(val, ":")
				cfg.Users[val] = &UserConfig{}

				// Initialize the new user
				cCfg = cfg.Users[val]
				cCfg.AllowCommands = cfg.AllowCommands
				cCfg.AllowEdit = cfg.AllowEdit
				cCfg.AllowNew = cfg.AllowEdit
				cCfg.Commands = cfg.Commands
				cCfg.FrontMatter = cfg.FrontMatter
				cCfg.PathScope = cfg.PathScope
				cCfg.Root = cfg.Root
				cCfg.Rules = cfg.Rules
				cCfg.StyleSheet = cfg.StyleSheet
			}
		}

		// Set global base url
		cfg.BaseURL = baseURL
		cfg.BaseURL = strings.TrimPrefix(cfg.BaseURL, "/")
		cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")
		cfg.BaseURL = "/" + cfg.BaseURL

		cfg.Root = http.Dir(cfg.PathScope)

		caddyConf := httpserver.GetConfig(c)
		cfg.AbsoluteURL = strings.TrimSuffix(caddyConf.Addr.Path, "/") + "/" + cfg.BaseURL
		cfg.AbsoluteURL = strings.Replace(cfg.AbsoluteURL, "//", "/", -1)
		cfg.AbsoluteURL = strings.TrimSuffix(cfg.AbsoluteURL, "/")
		cfg.AddrPath = strings.TrimSuffix(caddyConf.Addr.Path, "/")
		if err := appendConfig(cfg); err != nil {
			return configs, err
		}

	}

	return configs, nil
}

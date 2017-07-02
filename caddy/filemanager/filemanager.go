// Package filemanager provides middleware for managing files in a directory
// when directory path is requested instead of a specific file. Based on browse
// middleware.
package filemanager

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	. "github.com/hacdias/filemanager"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("filemanager", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

type plugin struct {
	Next    httpserver.Handler
	Configs []*config
}

type config struct {
	*FileManager
	baseURL   string
	webDavURL string
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
func (f plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	for i := range f.Configs {
		// Checks if this Path should be handled by File Manager.
		if !httpserver.Path(r.URL.Path).Matches(f.Configs[i].baseURL) {
			continue
		}

		return f.Configs[i].ServeHTTP(w, r)
	}

	return f.Next.ServeHTTP(w, r)
}

// setup configures a new FileManager middleware instance.
func setup(c *caddy.Controller) error {
	configs, err := parse(c)
	if err != nil {
		return err
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return plugin{Configs: configs, Next: next}
	})

	return nil
}

func parse(c *caddy.Controller) ([]*config, error) {
	var (
		configs []*config
		err     error
	)

	for c.Next() {
		var (
			m    = &config{FileManager: New(".")}
			u    = m.User
			name = ""
		)

		caddyConf := httpserver.GetConfig(c)

		m.SetPrefixURL(strings.TrimSuffix(caddyConf.Addr.Path, "/"))
		m.Commands = []string{"git", "svn", "hg"}
		m.Rules = append(m.Rules, &Rule{
			Regex:  true,
			Allow:  false,
			Regexp: regexp.MustCompile("\\/\\..+"),
		})

		// Get the baseURL
		args := c.RemainingArgs()

		if len(args) > 0 {
			m.baseURL = args[0]
			m.SetBaseURL(args[0])
		}

		for c.NextBlock() {
			switch c.Val() {
			case "before_save":
				if m.BeforeSave, err = makeCommand(c, m); err != nil {
					return configs, err
				}
			case "after_save":
				if m.AfterSave, err = makeCommand(c, m); err != nil {
					return configs, err
				}
			case "show":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				m.SetScope(c.Val(), name)
			case "styles":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				var tplBytes []byte
				tplBytes, err = ioutil.ReadFile(c.Val())
				if err != nil {
					return configs, err
				}

				u.StyleSheet = string(tplBytes)
			case "allow_new":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				u.AllowNew, err = strconv.ParseBool(c.Val())
				if err != nil {
					return configs, err
				}
			case "allow_edit":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				u.AllowEdit, err = strconv.ParseBool(c.Val())
				if err != nil {
					return configs, err
				}
			case "allow_commands":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				u.AllowCommands, err = strconv.ParseBool(c.Val())
				if err != nil {
					return configs, err
				}
			case "allow_command":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				u.Commands = append(u.Commands, c.Val())
			case "block_command":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				index := 0

				for i, val := range u.Commands {
					if val == c.Val() {
						index = i
					}
				}

				u.Commands = append(u.Commands[:index], u.Commands[index+1:]...)
			case "allow", "allow_r", "block", "block_r":
				ruleType := c.Val()

				if !c.NextArg() {
					return configs, c.ArgErr()
				}

				if c.Val() == "dotfiles" && !strings.HasSuffix(ruleType, "_r") {
					ruleType += "_r"
				}

				rule := &Rule{
					Allow: ruleType == "allow" || ruleType == "allow_r",
					Regex: ruleType == "allow_r" || ruleType == "block_r",
				}

				if rule.Regex && c.Val() == "dotfiles" {
					rule.Regexp = regexp.MustCompile("\\/\\..+")
				} else if rule.Regex {
					rule.Regexp = regexp.MustCompile(c.Val())
				} else {
					rule.Path = c.Val()
				}

				u.Rules = append(u.Rules, rule)
			default:
				// Is it a new user? Is it?
				val := c.Val()

				// Checks if it's a new user!
				if !strings.HasSuffix(val, ":") {
					fmt.Println("Unknown option " + val)
				}

				// Get the username, sets the current user, and initializes it
				val = strings.TrimSuffix(val, ":")
				m.NewUser(val)
				name = val
			}
		}

		m.baseURL = strings.TrimSuffix(m.baseURL, "/")
		m.webDavURL = strings.TrimSuffix(m.webDavURL, "/")
		configs = append(configs, m)
	}

	return configs, nil
}

func makeCommand(c *caddy.Controller, m *config) (Command, error) {
	fn := func(r *http.Request, c *FileManager, u *User) error { return nil }

	args := c.RemainingArgs()
	if len(args) == 0 {
		return fn, c.ArgErr()
	}

	nonblock := false
	if len(args) > 1 && args[len(args)-1] == "&" {
		// Run command in background; non-blocking
		nonblock = true
		args = args[:len(args)-1]
	}

	command, args, err := caddy.SplitCommandAndArgs(strings.Join(args, " "))
	if err != nil {
		return fn, c.Err(err.Error())
	}

	fn = func(r *http.Request, c *FileManager, u *User) error {
		path := strings.Replace(r.URL.Path, m.baseURL+m.webDavURL, "", 1)
		path = u.Scope + "/" + path
		path = filepath.Clean(path)

		for i := range args {
			args[i] = strings.Replace(args[i], "{path}", path, -1)
		}

		cmd := exec.Command(command, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if nonblock {
			log.Printf("[INFO] Nonblocking Command:\"%s %s\"", command, strings.Join(args, " "))
			return cmd.Start()
		}

		log.Printf("[INFO] Blocking Command:\"%s %s\"", command, strings.Join(args, " "))
		return cmd.Run()
	}

	return fn, nil
}

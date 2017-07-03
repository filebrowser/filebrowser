// Package filemanager provides middleware for managing files in a directory
// when directory path is requested instead of a specific file. Based on browse
// middleware.
package filemanager

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/net/webdav"

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
	baseURL string
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
	)

	for c.Next() {
		// TODO:
		// filemanager [baseurl] [baseScope] {
		//     database	path
		// }

		baseURL := "/"
		baseScope := "."

		// Get the baseURL and baseScope
		args := c.RemainingArgs()

		if len(args) == 1 {
			baseURL = args[0]
		}

		if len(args) > 1 {
			baseScope = args[1]
		}

		fm, err := New("./this.db", User{
			Username:      "admin",
			Password:      "admin",
			AllowCommands: true,
			AllowEdit:     true,
			AllowNew:      true,
			Commands:      []string{"git", "svn", "hg"},
			Rules: []*Rule{{
				Regex:  true,
				Allow:  false,
				Regexp: &Regexp{Raw: "\\/\\..+"},
			}},
			CSS:        "",
			FileSystem: webdav.Dir(baseScope),
		})

		if err != nil {
			return nil, err
		}

		caddyConf := httpserver.GetConfig(c)

		m := &config{FileManager: fm}
		m.SetBaseURL(baseURL)
		m.SetPrefixURL(strings.TrimSuffix(caddyConf.Addr.Path, "/"))
		m.baseURL = strings.TrimSuffix(baseURL, "/")

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
		path := strings.Replace(r.URL.Path, m.baseURL+"/files", "", 1)
		path = string(u.FileSystem) + "/" + path
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

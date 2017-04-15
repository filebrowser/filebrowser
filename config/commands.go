package config

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mholt/caddy"
)

// CommandFunc ...
type CommandFunc func(r *http.Request, c *Config, u *User) error

// CommandRunner ...
func CommandRunner(c *caddy.Controller) (CommandFunc, error) {
	fn := func(r *http.Request, c *Config, u *User) error { return nil }

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

	fn = func(r *http.Request, c *Config, u *User) error {
		path := strings.Replace(r.URL.Path, c.WebDavURL, "", 1)
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

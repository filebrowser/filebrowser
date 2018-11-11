package types

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/mholt/caddy"
)

// DefaultEvents are the default events that can trigger commands.
// See Settings.Commands.
var DefaultEvents = []string{
	"before_save",
	"after_save",
	"before_publish",
	"after_publish",
	"before_copy",
	"after_copy",
	"before_rename",
	"after_rename",
	"before_upload",
	"after_upload",
	"before_delete",
	"after_delete",
}

// Runner runs certain commands.
type Runner struct {
	Commands map[string][]string `json:"commands"`
}

// Run runs the hooks for the before and after event.
func (r Runner) Run(fn func() error, event string, path string, dst string, u *User) error {
	err := r.do("before_"+event, path, dst, u)
	if err != nil {
		return err
	}

	err = fn()
	if err != nil {
		return err
	}

	return r.do("after_"+event, path, dst, u)
}

func (r Runner) do(event string, path string, destination string, user *User) error {
	commands := []string{}

	if val, ok := r.Commands[event]; ok {
		commands = append(commands, val...)
	}

	for _, command := range commands {
		args := strings.Split(command, " ")
		nonblock := false

		if len(args) > 1 && args[len(args)-1] == "&" {
			nonblock = true
			args = args[:len(args)-1]
		}

		command, args, err := caddy.SplitCommandAndArgs(strings.Join(args, " "))
		if err != nil {
			return err
		}

		cmd := exec.Command(command, args...)
		cmd.Env = append(os.Environ(), fmt.Sprintf("FILE=%s", path))
		cmd.Env = append(cmd.Env, fmt.Sprintf("ROOT=%s", user.Scope))
		cmd.Env = append(cmd.Env, fmt.Sprintf("TRIGGER=%s", event))
		cmd.Env = append(cmd.Env, fmt.Sprintf("USERNAME=%s", user.Username))

		if destination != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("DESTINATION=%s", destination))
		}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if nonblock {
			log.Printf("[INFO] Nonblocking Command:\"%s %s\"", command, strings.Join(args, " "))
			if err := cmd.Start(); err != nil {
				return err
			}

			continue
		}

		log.Printf("[INFO] Blocking Command:\"%s %s\"", command, strings.Join(args, " "))
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

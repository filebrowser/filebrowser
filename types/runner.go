package types

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/mholt/caddy"
)

var defaultEvents = []string{
	"save",
	"copy",
	"rename",
	"upload",
	"delete",
}

// Run runs the hooks for the before and after event.
func (s *Settings) Run(fn func() error, evt, path, dst string, user *User) error {
	path = user.FullPath(path)
	dst = user.FullPath(dst)

	if val, ok := s.Commands["before_"+evt]; ok {
		for _, command := range val {
			err := s.exec(command, "before_"+evt, path, dst, user)
			if err != nil {
				return err
			}
		}
	}

	err := fn()
	if err != nil {
		return err
	}

	if val, ok := s.Commands["after_"+evt]; ok {
		for _, command := range val {
			err := s.exec(command, "after_"+evt, path, dst, user)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ParseCommand parses the command taking in account
func (s *Settings) ParseCommand(raw string) ([]string, error) {
	command := []string{}

	if len(s.Shell) == 0 {
		cmd, args, err := caddy.SplitCommandAndArgs(raw)
		if err != nil {
			return nil, err
		}

		_, err = exec.LookPath(cmd)
		if err != nil {
			return nil, err
		}

		command = append(command, cmd)
		command = append(command, args...)
	} else {
		command = append(s.Shell, raw)
	}

	return command, nil
}

func (s *Settings) exec(raw, evt, path, dst string, user *User) error {
	blocking := true

	if strings.HasSuffix(raw, "&") {
		blocking = false
		raw = strings.TrimSpace(strings.TrimSuffix(raw, "&"))
	}

	command, err := s.ParseCommand(raw)
	if err != nil {
		return err
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("FILE=%s", path))
	cmd.Env = append(cmd.Env, fmt.Sprintf("SCOPE=%s", user.Scope))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TRIGGER=%s", evt))
	cmd.Env = append(cmd.Env, fmt.Sprintf("USERNAME=%s", user.Username))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DESTINATION=%s", dst))

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if !blocking {
		log.Printf("[INFO] Nonblocking Command: \"%s\"", strings.Join(command, " "))
		return cmd.Start()
	}

	log.Printf("[INFO] Blocking Command: \"%s\"", strings.Join(command, " "))
	return cmd.Run()
}

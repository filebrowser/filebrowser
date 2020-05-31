package runner

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// Runner is a commands runner.
type Runner struct {
	*settings.Settings
}

// RunHook runs the hooks for the before and after event.
func (r *Runner) RunHook(fn func() error, evt, path, dst string, user *users.User) error {
	path = user.FullPath(path)
	dst = user.FullPath(dst)

	if val, ok := r.Commands["before_"+evt]; ok {
		for _, command := range val {
			err := r.exec(command, "before_"+evt, path, dst, user)
			if err != nil {
				return err
			}
		}
	}

	err := fn()
	if err != nil {
		return err
	}

	if val, ok := r.Commands["after_"+evt]; ok {
		for _, command := range val {
			err := r.exec(command, "after_"+evt, path, dst, user)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Runner) exec(raw, evt, path, dst string, user *users.User) error {
	blocking := true

	if strings.HasSuffix(raw, "&") {
		blocking = false
		raw = strings.TrimSpace(strings.TrimSuffix(raw, "&"))
	}

	command, err := ParseCommand(r.Settings, raw)
	if err != nil {
		return err
	}

	cmd := exec.Command(command[0], command[1:]...) //nolint:gosec
	cmd.Env = append(os.Environ(), fmt.Sprintf("FILE=%s", path))
	cmd.Env = append(cmd.Env, fmt.Sprintf("SCOPE=%s", user.Scope)) //nolint:gocritic
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

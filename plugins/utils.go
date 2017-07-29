package plugins

import (
	"errors"
	"os/exec"
)

// Run executes an external command
func Run(command string, args []string, path string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = path
	out, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New(string(out))
	}

	return nil
}

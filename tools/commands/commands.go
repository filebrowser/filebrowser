package commands

import (
	"os"
	"os/exec"
)

// RunCommand executes an external command
func Run(command string, args []string, path string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = path
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

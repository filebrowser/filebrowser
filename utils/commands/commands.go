package commands

import (
	"os"
	"os/exec"
)

// Run executes an external command
func Run(command string, args []string, path string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

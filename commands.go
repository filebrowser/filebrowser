package hugo

import (
	"log"
	"os"
	"os/exec"
)

// Run executes an external command
func Run(command string, args []string, path string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = path
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunHugo is used to run the static website generator
func RunHugo(c *Config, force bool) {
	os.RemoveAll(c.Public)

	// Prevent running if watching is enabled
	if b, pos := StringInSlice("--watch", c.Args); b && !force {
		if len(c.Args) > pos && c.Args[pos+1] != "false" {
			return
		}

		if len(c.Args) == pos+1 {
			return
		}
	}

	if err := Run(c.Hugo, c.Args, c.Root); err != nil {
		log.Panic(err)
	}
}

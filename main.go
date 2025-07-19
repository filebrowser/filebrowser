package main

import (
	"os"

	"github.com/filebrowser/filebrowser/v2/cmd"
	"github.com/filebrowser/filebrowser/v2/errors"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(errors.GetExitCode(err))
	}
}

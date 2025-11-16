package main

import (
	"os"

	"github.com/filebrowser/filebrowser/v2/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

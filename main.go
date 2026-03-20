package main

import (
	"os"

	"github.com/rforced/filebrowser/v2/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

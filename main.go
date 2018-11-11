package main

import (
	"runtime"

	"github.com/filebrowser/filebrowser/cmd"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()
}

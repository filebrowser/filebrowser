//go:generate cd http && rice embed-go
package main

import (
	"runtime"

	"github.com/filebrowser/filebrowser/v2/cmd"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()
}

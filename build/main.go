package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	goPath := os.Getenv("GOPATH")
	hugoPath := filepath.Join(goPath, "src/github.com/spf13/hugo")

	if found, err := exists(hugoPath); !found || err != nil {
		log.Fatalf("Aborting. Can't find Hugo source on %s.", hugoPath)
	}

	// NOTE: I assume that 'go get -u' was run before of this and that
	// every package and dependency is up to date.

	// Get new tags from remote
	run("git", []string{"fetch", "--tags"}, hugoPath)

	// Get the revision for the latest tag
	commit := run("git", []string{"rev-list", "--tags", "--max-count=1"}, hugoPath)

	// Get the latest tag
	tag := run("git", []string{"describe", "--tags", commit}, hugoPath)

	// Checkout the latest tag
	run("git", []string{"checkout", tag}, hugoPath)

	// Build hugo binary
	pluginPath := filepath.Join(goPath, "src/github.com/hacdias/caddy-hugo")
	run("go", []string{"build", "-o", "assets/hugo", "github.com/spf13/hugo"}, pluginPath)
}

func run(command string, args []string, path string) string {
	cmd := exec.Command(command, args...)
	cmd.Dir = path
	out, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(string(out))
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

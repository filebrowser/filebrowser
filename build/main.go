package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Getenv("TRAVIS")) > 0 || len(os.Getenv("CI")) > 0 {
		return
	}

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

	updateVersion(pluginPath, tag)
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

func updateVersion(path string, version string) {
	path = filepath.Join(path, "installer.go")

	input, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "const version") {
			lines[i] = "const version = \"" + version + "\""
		}
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(path, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

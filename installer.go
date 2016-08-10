package hugo

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"

	"github.com/mitchellh/go-homedir"
)

// This is automatically set on `go generate`
const version = "UNDEFINED"

// GetPath retrives the Hugo path for the user or install it if it's not found
func getPath() string {
	homedir, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	caddy := filepath.Join(homedir, ".caddy")
	bin := filepath.Join(caddy, "bin")
	hugo := ""
	found := false

	// Check if Hugo is already on $PATH
	if hugo, err = exec.LookPath("hugo"); err == nil {
		if checkVersion(hugo) {
			return hugo
		}

		found = true
	}

	if !found {
		hugo = filepath.Join(bin, "hugo")

		if runtime.GOOS == "windows" {
			hugo += ".exe"
		}

		// Check if Hugo is on $HOME/.caddy/bin
		if _, err = os.Stat(hugo); err == nil {
			if checkVersion(hugo) {
				return hugo
			}

			found = true
		}
	}

	if found {
		fmt.Println("We will update your hugo to the newest version.")
	} else {
		fmt.Println("Unable to find Hugo on your computer.")
	}

	// Create the neccessary folders
	os.MkdirAll(caddy, 0774)
	os.Mkdir(bin, 0774)

	binary, err := Asset("hugo")

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	err = ioutil.WriteFile(hugo, binary, 0644)

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	binary = nil

	// Force memory RAM garbage collector
	debug.FreeOSMemory()

	fmt.Println("Hugo installed at " + hugo)
	return hugo
}

func checkVersion(hugo string) bool {
	out, _ := exec.Command(hugo, "version").Output()

	r := regexp.MustCompile(`v\d\.\d{2}`)
	v := r.FindStringSubmatch(string(out))[0]

	return (v == version)
}

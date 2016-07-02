package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/hacdias/caddy-hugo/utils/files"
	"github.com/mitchellh/go-homedir"
	"github.com/pivotal-golang/archiver/extractor"
)

const (
	version = "0.16"
	baseurl = "https://github.com/spf13/hugo/releases/download/v" + version + "/"
)

var caddy, bin, temp, hugo, tempfile, zipname, exename string

// GetPath retrives the Hugo path for the user or install it if it's not found
func GetPath() string {
	initializeVariables()

	var err error
	found := false

	// Check if Hugo is already on $PATH
	if hugo, err = exec.LookPath("hugo"); err == nil {
		if checkVersion() {
			return hugo
		}

		found = true
	}

	// Check if Hugo is on $HOME/.caddy/bin
	if _, err = os.Stat(hugo); err == nil {
		if checkVersion() {
			return hugo
		}

		found = true
	}

	if found {
		fmt.Println("We will update your hugo to the newest version.")
	} else {
		fmt.Println("Unable to find Hugo on your computer.")
	}

	// Create the neccessary folders
	os.MkdirAll(caddy, 0774)
	os.Mkdir(bin, 0774)

	if temp, err = ioutil.TempDir("", "caddy-hugo"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	downloadHugo()
	checkSHA256()

	fmt.Print("Unzipping... ")

	// Unzip or Ungzip the file
	switch runtime.GOOS {
	case "windows":
		zp := extractor.NewZip()
		err = zp.Extract(tempfile, temp)
	default:
		gz := extractor.NewTgz()
		err = gz.Extract(tempfile, temp)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	fmt.Println("done.")

	var exetorename string

	err = filepath.Walk(temp, func(path string, f os.FileInfo, err error) error {
		if f.Name() == exename {
			exetorename = path
		}

		return nil
	})

	// Copy the file
	fmt.Print("Moving Hugo executable... ")
	err = files.CopyFile(exetorename, hugo)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	err = os.Chmod(hugo, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	fmt.Println("done.")
	fmt.Println("Hugo installed at " + hugo)
	defer os.RemoveAll(temp)
	return hugo
}

func initializeVariables() {
	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "64bit"
	case "386":
		arch = "32bit"
	case "arm":
		arch = "arm32"
	default:
		arch = runtime.GOARCH
	}

	var ops = runtime.GOOS
	if runtime.GOOS == "darwin" && runtime.GOARCH != "arm" {
		ops = "osx"
	}

	exename = "hugo"
	zipname = "hugo_" + version + "_" + ops + "-" + arch

	homedir, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	caddy = filepath.Join(homedir, ".caddy")
	bin = filepath.Join(caddy, "bin")
	hugo = filepath.Join(bin, "hugo")

	switch runtime.GOOS {
	case "windows":
		zipname += ".zip"
		exename += ".exe"
		hugo += ".exe"
	default:
		zipname += ".tgz"
	}
}

func checkVersion() bool {
	out, _ := exec.Command("hugo", "version").Output()

	r := regexp.MustCompile(`v\d\.\d{2}`)
	v := r.FindStringSubmatch(string(out))[0]
	v = v[1:len(v)]

	return (v == version)
}

func downloadHugo() {
	tempfile = filepath.Join(temp, zipname)

	fmt.Print("Downloading Hugo from GitHub releases... ")

	// Create the file
	out, err := os.Create(tempfile)
	out.Chmod(0774)
	if err != nil {
		defer os.RemoveAll(temp)
		fmt.Println(err)
		os.Exit(-1)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(baseurl + zipname)
	if err != nil {
		fmt.Println("An error ocurred while downloading. If this error persists, try downloading Hugo from \"https://github.com/spf13/hugo/releases/\" and put the executable in " + bin + " and rename it to 'hugo' or 'hugo.exe' if you're on Windows.")
		fmt.Println(err)
		os.Exit(-1)
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	fmt.Println("downloaded.")
}

func checkSHA256() {
	fmt.Print("Checking SHA256...")

	hasher := sha256.New()
	f, err := os.Open(tempfile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := io.Copy(hasher, f); err != nil {
		log.Fatal(err)
	}

	if hex.EncodeToString(hasher.Sum(nil)) != sha256Hash[zipname] {
		fmt.Println("can't verify SHA256.")
		os.Exit(-1)
	}

	fmt.Println("checked!")
}

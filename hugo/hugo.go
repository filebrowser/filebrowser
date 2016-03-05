package hugo

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
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/pivotal-golang/archiver/extractor"
)

const (
	version = "0.15"
	baseurl = "https://github.com/spf13/hugo/releases/download/v" + version + "/"
)

var (
	caddy, bin, temp, hugo, tempfile, zipname, exename string
	sha256Hash                                         = map[string]string{
		"hugo_0.15_darwin_386.zip":              "f9b7353f9b64e7aece5f7981e5aa97dc4b31974ce76251edc070e77691bc03e2",
		"hugo_0.15_darwin_amd64.zip":            "aeecd6a12d86ab920f5b04e9486474bbe478dc246cdc2242799849b84c61c6f1",
		"hugo_0.15_dragonfly_amd64.zip":         "e380343789f2b2e0c366c8e1eeb251ccd90eea53dac191ff85d8177b130e53bc",
		"hugo_0.15_freebsd_386.zip":             "98f9210bfa3dcb48bd154879ea1cfe1b0ed8a3d891fdeacbdb4c3fc69b72aac4",
		"hugo_0.15_freebsd_amd64.zip":           "aa6a3028899e76e6920b9b5a64c29e14017ae34120efa67276e614e3a69cb100",
		"hugo_0.15_freebsd_arm.zip":             "de52e1b07caf778bdc3bdb07f39119cd5a1739c8822ebe311cd4f667c43588ac",
		"hugo_0.15_linux_386.tar.gz":            "af28c4cbb16db765535113f361a38b2249c634ce2d3798dcf5b795de6e4b7ecf",
		"hugo_0.15_linux_amd64.tar.gz":          "32a6335bd76f72867efdec9306a8a7eb7b9498a2e0478105efa96c1febadb09b",
		"hugo_0.15_linux_arm.tar.gz":            "886dd1a843c057a46c541011183dd558469250580e81450eedbd1a4d041e9234",
		"hugo_0.15_netbsd_386.zip":              "6245f5db16b33a09466f149d5b7b68a7899d6d624903de9e7e70c4b6ea869a72",
		"hugo_0.15_netbsd_amd64.zip":            "103ea8d81d2a3d707c05e3dd68c98fcf8146ddd36b49bf0e65d9874cee230c88",
		"hugo_0.15_netbsd_arm.zip":              "9c9b5cf4ea3b6169be1b5fc924251a247d9c140dd8a45aa5175031878585ff0a",
		"hugo_0.15_openbsd_386.zip":             "81dfdb3048a27a61b249650241fe4e8da1eda31a3a7311c615eb419f1cdd06b1",
		"hugo_0.15_openbsd_amd64.zip":           "e7447cde0dd7628b05b25b86938018774d8db8156ab1330b364e0e2c6501ad87",
		"hugo_0.15_windows_386_32-bit-only.zip": "0a72f9a1a929f36c0e52fb1c6272b4d37a2bd1a6bd19ce57a6e7b6803b434756",
		"hugo_0.15_windows_amd64.zip":           "9f03602e48ae2199e06431d7436fb3b9464538c0d44aac9a76eb98e1d4d5d727",
	}
)

// GetPath retrives the Hugo path for the user or install it if it's not found
func GetPath() string {
	initializeVariables()

	var err error

	// Check if Hugo is already on $PATH
	if hugo, err := exec.LookPath("hugo"); err == nil {
		return hugo
	}

	// Check if Hugo is on $HOME/.caddy/bin
	if _, err = os.Stat(hugo); err == nil {
		return hugo
	}

	fmt.Println("Unable to find Hugo on your computer.")

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
	case "darwin", "windows":
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

	err = os.Rename(exetorename, hugo)

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	fmt.Println("Hugo installed at " + hugo)
	defer os.RemoveAll(temp)
	return hugo
}

func initializeVariables() {
	exename = "hugo_" + version + "_" + runtime.GOOS + "_" + runtime.GOARCH
	zipname = exename

	homedir, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	caddy = filepath.Join(homedir, ".caddy")
	bin = filepath.Join(caddy, "bin")
	hugo = filepath.Join(bin, "hugo")

	switch runtime.GOOS {
	case "darwin":
		zipname += ".zip"
	case "windows":
		// At least for v0.15 version
		if runtime.GOARCH == "386" {
			zipname += "32-bit-only"
		}

		zipname += ".zip"
		exename += ".exe"
		hugo += ".exe"
	default:
		zipname += ".tar.gz"
	}
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

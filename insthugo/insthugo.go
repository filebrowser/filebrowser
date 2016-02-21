package insthugo

import (
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	version = "0.15"
	baseurl = "https://github.com/spf13/hugo/releases/download/v" + version + "/"
)

var (
	usr        user.User
	tempfiles  []string
	filename   = "hugo_" + version + "_" + runtime.GOOS + "_" + runtime.GOARCH
	sha256Hash = map[string]string{
		"hugo_0.15_darwin_386.zip":              "",
		"hugo_0.15_darwin_amd64.zip":            "",
		"hugo_0.15_dragonfly_amd64.zip":         "",
		"hugo_0.15_freebsd_386.zip":             "",
		"hugo_0.15_freebsd_amd64.zip":           "",
		"hugo_0.15_freebsd_arm.zip":             "",
		"hugo_0.15_linux_386.tar.gz":            "",
		"hugo_0.15_linux_amd64.tar.gz":          "",
		"hugo_0.15_linux_arm.tar.gz":            "",
		"hugo_0.15_netbsd_386.zip":              "",
		"hugo_0.15_netbsd_amd64.zip":            "",
		"hugo_0.15_netbsd_arm.zip":              "",
		"hugo_0.15_openbsd_386.zip":             "",
		"hugo_0.15_openbsd_amd64.zip":           "",
		"hugo_0.15_windows_386_32-bit-only.zip": "0a72f9a1a929f36c0e52fb1c6272b4d37a2bd1a6bd19ce57a6e7b6803b434756",
		"hugo_0.15_windows_amd64.zip":           "9f03602e48ae2199e06431d7436fb3b9464538c0d44aac9a76eb98e1d4d5d727",
	}
)

// Install installs Hugo
func Install() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	caddy := filepath.Clean(usr.HomeDir + "/.caddy/")
	bin := filepath.Clean(caddy + "/bin")
	temp := filepath.Clean(caddy + "/temp")
	hugo := filepath.Clean(bin + "/hugo")

	switch runtime.GOOS {
	case "darwin":
		filename += ".zip"
	case "windows":
		// At least for v0.15 version
		if runtime.GOARCH == "386" {
			filename += "32-bit-only"
		}

		filename += ".zip"
		hugo += ".exe"
	default:
		filename += ".tar.gz"
	}

	// Check if Hugo is already installed
	if _, err := os.Stat(hugo); err == nil {
		return hugo
	}

	fmt.Println("Unable to find Hugo on " + caddy)

	err = os.MkdirAll(caddy, 0666)
	err = os.Mkdir(bin, 0666)
	err = os.Mkdir(temp, 0666)

	if !os.IsExist(err) {
		fmt.Println(err)
		os.Exit(-1)
	}

	tempfile := temp + "/" + filename

	// Create the file
	tempfiles = append(tempfiles, tempfile)
	out, err := os.Create(tempfile)
	if err != nil {
		clean()
		fmt.Println(err)
		os.Exit(-1)
	}
	defer out.Close()

	fmt.Print("Downloading Hugo from GitHub releases... ")

	// Get the data
	resp, err := http.Get(baseurl + filename)
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

	if hex.EncodeToString(hasher.Sum(nil)) != sha256Hash[filename] {
		fmt.Println("can't verify SHA256.")
		os.Exit(-1)
	}

	fmt.Println("checked!")
	fmt.Print("Unziping... ")

	// Unzip or Ungzip the file
	switch runtime.GOOS {
	case "darwin", "windows":
		err = unzip(tempfile, bin)
	default:
		err = ungzip(tempfile, bin)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	fmt.Println("done.")

	tempfiles = append(tempfiles, bin+"README.md", bin+"LICENSE.md")
	clean()

	ftorename := bin + strings.Replace(filename, ".tar.gz", "", 1)

	if runtime.GOOS == "windows" {
		ftorename = bin + strings.Replace(filename, ".zip", ".exe", 1)
	}

	os.Rename(ftorename, hugo)
	fmt.Println("Hugo installed at " + hugo)
	return hugo
}

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}

func ungzip(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}

func clean() {
	fmt.Print("Removing temporary files... ")

	for _, file := range tempfiles {
		os.Remove(file)
	}

	fmt.Println("done.")
}

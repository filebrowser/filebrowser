package insthugo

import (
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

const version = "0.15"

// Install installs Hugo
func Install() error {
	// Sets the base url from where to download
	baseurl := "https://github.com/spf13/hugo/releases/download/v" + version + "/"

	// The default filename
	filename := "hugo_" + version + "_" + runtime.GOOS + "_" + runtime.GOARCH

	switch runtime.GOOS {
	case "darwin", "windows":
		// At least for v0.15 version
		if runtime.GOOS == "windows" && runtime.GOARCH == "386" {
			filename += "32-bit-only"
		}

		filename += ".zip"
	default:
		filename += ".tar.gz"
	}

	// Gets the current user home directory and creates the .caddy dir
	user, err := user.Current()
	if err != nil {
		return err
	}

	path := user.HomeDir + "/.caddy/"
	bin := path + "bin/"
	temp := path + "temp/"

	err = os.MkdirAll(path, 0666)
	err = os.Mkdir(bin, 0666)
	err = os.Mkdir(temp, 0666)
	if err != nil {
		return err
	}

	tempfile := temp + filename

	// Create the file
	out, err := os.Create(tempfile)
	if err != nil {
		return err
	}
	defer out.Close()

	fmt.Println("Downloading Hugo...")

	// Get the data
	resp, err := http.Get(baseurl + filename)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("Checking SHA256...")
	// TODO: check sha256

	fmt.Println("Unziping...")

	// Unzip or Ungzip the file
	switch runtime.GOOS {
	case "darwin", "windows":
		unzip(temp+filename, bin)
	default:
		ungzip(temp+filename, bin)
	}

	fmt.Println("Removing temporary files...")

	// Removes the temporary file and other files
	os.Remove(tempfile)
	os.Remove(bin + "README.md")
	os.Remove(bin + "LICENSE.md")

	if runtime.GOOS == "windows" {
		os.Rename(bin+strings.Replace(filename, ".zip", ".exe", 1), bin+"hugo.exe")

		fmt.Println("Hugo installed at " + filepath.Clean(bin) + "\\hugo.exe")
		return nil
	}

	os.Rename(bin+strings.Replace(filename, ".tar.gz", "", 1), bin+"hugo")
	fmt.Println("Hugo installed at " + filepath.Clean(bin) + "/hugo")
	return nil
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

package downloader

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Wget struct {
	URL      string `json:"url,omitempty"`
	Filename string `json:"filename,omitempty"`
	Pathname string `json:"pathname,omitempty"`
	total    int64
	received int64
}

func newWget(url string, filename string, pathname string) *Wget {
	return &Wget{
		URL:      url,
		Filename: filename,
		Pathname: pathname,
	}
}

func (w *Wget) Download(url string, filename string, pathname string) error {
	_, err := os.Stat(pathname)
	if err != nil && os.IsNotExist(err) {
		err := os.Mkdir(pathname, 0755)
		if err != nil {
			return err
		}
	}
	downloadFilepath := filepath.Join(pathname, filename)
	_, err = os.Stat(downloadFilepath)
	if err != nil && os.IsExist(err) {
		return err
	}
	output, err := exec.Command("wget", "-O", downloadFilepath, url).CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", output)
	return nil
}

func (w *Wget) GetRatio() float64 {
	return float64(w.received) / float64(w.total)
}

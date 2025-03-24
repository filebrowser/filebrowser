package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type Downloader interface {
	Download(url string, filename string, pathname string) error
	GetRatio() float64
}

type DownloadTask struct {
	Filename string
	Path     string
	Url      string
}

func NewDownloadTask(filename string, path string, url string) *DownloadTask {
	return &DownloadTask{
		Filename: filename,
		Path:     path,
		Url:      url,
	}
}

func (t *DownloadTask) Valid() error {
	if t.Filename == "" {
		return errors.New("filename is empty")
	}
	if t.Url == "" {
		return errors.New("url is empty")
	}
	if t.Path == "" {
		return errors.New("path is empty")
	}
	if strings.Contains(t.Path, ",,") {
		return errors.New("path is invalid")
	}
	return nil
}

func isExists(pathname string) bool {
	_, err := os.Stat(pathname)
	return err == nil || !os.IsNotExist(err)
}

func isDir(pathname string) bool {
	info, err := os.Stat(pathname)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (t *DownloadTask) Download() error {
	if err := t.Valid(); err != nil {
		return err
	}
	if isExists(t.Path) {
		if !isDir(t.Path) {
			return errors.New("path is not a directory")
		}
		if isExists(path.Join(t.Path, t.Filename)) {
			return errors.New("file already exists")
		}
	} else {
		if err := os.Mkdir(t.Path, 0755); err != nil {
			return err
		}
	}
	resp, err := http.Get(t.Url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", t.Url, err)
	}
	defer resp.Body.Close()
	contentLength := resp.ContentLength
	if contentLength < 0 {
		return errors.New("failed to get content length")
	}
	file, err := os.Create(path.Join(t.Path, t.Filename))
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", t.Filename, err)
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", t.Filename, err)
	}
	return nil
}

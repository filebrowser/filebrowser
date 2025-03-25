package http

import (
	"encoding/json"
	"fmt"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/afero"
	"io"
	"net/http"
	"os"
	"path"
)

type DownloadTask struct {
	TaskID    uuid.UUID `json:"taskID"`
	URL       string    `json:"url"`
	Filename  string    `json:"filename"`
	Pathname  string    `json:"pathname"`
	totalSize int64
	savedSize int64
	cache     *cache.Cache
}

func (d *DownloadTask) Progress() float64 {
	if d.totalSize == 0 {
		return 0
	}
	return float64(d.savedSize) / float64(d.totalSize)
}

func NewDownloadTask(url, filename, pathname string, downloaderCache *cache.Cache) *DownloadTask {
	taskId := uuid.New()
	downloadTask := &DownloadTask{
		TaskID:   taskId,
		URL:      url,
		Filename: filename,
		Pathname: pathname,
		cache:    downloaderCache,
	}
	downloaderCache.Set(taskId.String(), downloadTask, cache.NoExpiration)
	return downloadTask
}

type WriteCounter struct {
	task *DownloadTask
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.task.savedSize += int64(n)
	wc.task.cache.Set(wc.task.TaskID.String(), wc.task, cache.NoExpiration)
	fmt.Printf("Downloaded %d of %d bytes, percent: %.2f%%\n", wc.task.savedSize, wc.task.totalSize, wc.task.Progress())
	return n, nil
}

func downloadHandler(downloaderCache *cache.Cache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Create || !d.Check(r.URL.Path) {
			return http.StatusForbidden, nil
		}
		var params struct {
			URL      string `json:"url"`
			Filename string `json:"filename"`
			Pathname string `json:"pathname"`
		}
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			return http.StatusBadRequest, err
		}

		_, err := os.Stat(path.Join(params.Pathname, params.Filename))
		if err != nil && !os.IsNotExist(err) {
			return errToStatus(err), err
		}
		downloadTask := NewDownloadTask(params.URL, params.Filename, params.Pathname, downloaderCache)
		asyncDownloadWithTask(d.user.Fs, downloadTask)

		_, err = w.Write([]byte(downloadTask.TaskID.String()))
		if err != nil {
			return errToStatus(err), err
		}
		return 0, nil
	})
}

func downloadWithTask(fs afero.Fs, task *DownloadTask) error {
	err := fs.MkdirAll(task.Pathname, files.PermDir)
	if err != nil {
		return err
	}

	file, err := fs.OpenFile(path.Join(task.Pathname, task.Filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, files.PermFile)
	if err != nil {
		return err
	}
	defer file.Close()
	resp, err := http.Get(task.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	task.totalSize = resp.ContentLength

	_, err = io.Copy(file, io.TeeReader(resp.Body, &WriteCounter{task: task}))
	if err != nil {
		return err
	}

	return nil
}

func asyncDownloadWithTask(fs afero.Fs, task *DownloadTask) {
	go func() {
		err := downloadWithTask(fs, task)
		if err != nil {
			fmt.Printf("Error downloading file: %v\n", err)
			return
		}
	}()
}

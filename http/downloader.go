package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/afero"
	"io"
	"math"
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
	cancel    context.CancelFunc
	status    string
	err       error
}

func (d *DownloadTask) SaveCache() {
	d.cache.Set(d.TaskID.String(), d, cache.NoExpiration)
}

func (d *DownloadTask) Progress() float64 {
	if d.totalSize == 0 {
		return 0
	}
	return float64(d.savedSize) / float64(d.totalSize)
}

func (d *DownloadTask) ResolveErr(err error) {
	if err != nil {
		d.status = "error"
		d.err = err
		d.SaveCache()
	}
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
	return downloadTask
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

func downloadStatusHandler(downloaderCache *cache.Cache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Create || !d.Check(r.URL.Path) {
			return http.StatusForbidden, nil
		}
		taskID := r.URL.Path
		taskCacheRaw, ok := downloaderCache.Get(taskID)
		if !ok {
			return http.StatusNotFound, nil
		}
		taskCache := taskCacheRaw.(*DownloadTask)
		responseBody := map[string]interface{}{
			"progress":  math.Round(taskCache.Progress()*1000) / 1000,
			"totalSize": taskCache.totalSize,
			"savedSize": taskCache.savedSize,
			"filename":  taskCache.Filename,
			"pathname":  taskCache.Pathname,
			"url":       taskCache.URL,
			"taskID":    taskCache.TaskID.String(),
			"status":    taskCache.status,
			"error":     taskCache.err,
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(&responseBody)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		return 0, nil
	})
}

func downloadWithTask(fs afero.Fs, task *DownloadTask) error {
	ctx, cancel := context.WithCancel(context.Background())
	task.cancel = cancel
	task.status = "downloading"
	task.SaveCache()
	defer cancel()

	err := fs.MkdirAll(task.Pathname, files.PermDir)
	if err != nil {
		task.ResolveErr(err)
		return err
	}

	file, err := fs.OpenFile(path.Join(task.Pathname, task.Filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, files.PermFile)
	if err != nil {
		task.ResolveErr(err)
		return err
	}
	defer file.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, task.URL, nil)
	if err != nil {
		task.ResolveErr(err)
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		task.ResolveErr(err)
		return err
	}
	defer resp.Body.Close()
	task.totalSize = resp.ContentLength
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			task.status = "canceled"
			task.SaveCache()
			return ctx.Err()
		default:
			rn, err := resp.Body.Read(buf)
			if err == io.EOF {
				task.status = "completed"
				task.SaveCache()
				return nil
			}
			if err != nil {
				task.ResolveErr(err)
				return err
			}
			if rn > 0 {
				wn, err := file.Write(buf[:rn])
				if err != nil {
					task.ResolveErr(err)
					return err
				}
				task.savedSize += int64(wn)
				task.SaveCache()

			}
		}

	}
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

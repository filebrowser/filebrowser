package fbhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/filebrowser/filebrowser/v2/search"
)

const searchPingInterval = 5

var searchHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	response := make(chan map[string]interface{})
	ctx, cancel := context.WithCancelCause(r.Context())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Avoid connection timeout
		timeout := time.NewTimer(searchPingInterval * time.Second)
		defer timeout.Stop()
		for {
			var err error
			var infoBytes []byte
			select {
			case info := <-response:
				if info == nil {
					return
				}
				infoBytes, err = json.Marshal(info)
			case <-timeout.C:
				// Send a heartbeat packet
				infoBytes = nil
			case <-ctx.Done():
				return
			}
			if err != nil {
				cancel(err)
				return
			}
			_, err = w.Write(infoBytes)
			if err == nil {
				_, err = w.Write([]byte("\n"))
			}
			if err != nil {
				cancel(err)
				return
			}
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}()
	query := r.URL.Query().Get("query")

	err := search.Search(ctx, d.user.Fs, r.URL.Path, query, d, func(path string, f os.FileInfo) error {
		select {
		case <-ctx.Done():
		case response <- map[string]interface{}{
			"dir":  f.IsDir(),
			"path": path,
		}:
		}
		return context.Cause(ctx)
	})
	close(response)
	wg.Wait()
	if err == nil {
		err = context.Cause(ctx)
	}
	// ignore cancellation errors from user aborts
	if err != nil && !errors.Is(err, context.Canceled) {
		return http.StatusInternalServerError, err
	}

	return 0, nil
})

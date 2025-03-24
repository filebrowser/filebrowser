package http

import (
	"encoding/json"
	"github.com/filebrowser/filebrowser/v2/downloader"
	"net/http"
)

func downloadHandler() handleFunc {
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
		downloadTask := downloader.NewDownloadTask(params.Filename, params.Pathname, params.URL)

		err := downloadTask.Download()

		if err != nil {
			return http.StatusInternalServerError, err
		}
		return http.StatusNoContent, nil
	})
}

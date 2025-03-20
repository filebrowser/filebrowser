package http

import (
	"encoding/json"
	"fmt"
	"github.com/filebrowser/filebrowser/v2/downloader"
	"net/http"
)

func downloadHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		fmt.Printf("wget: %v\n", d.user.Perm.Create)
		if !d.user.Perm.Create {
			return http.StatusForbidden, nil
		}
		var wget downloader.Wget
		if err := json.NewDecoder(r.Body).Decode(&wget); err != nil {
			return http.StatusBadRequest, err
		}

		err := wget.Download(wget.URL, wget.Filename, wget.Pathname)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		return http.StatusNoContent, nil
	})
}

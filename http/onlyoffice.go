package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type OnlyOfficeCallback struct {
	ChangesURL string   `json:"changesurl,omitempty"`
	Key        string   `json:"key"`
	Status     int      `json:"status"`
	URL        string   `json:"url,omitempty"`
	Users      []string `json:"users,omitempty"`
	UserData   string   `json:"userdata,omitempty"`
}

var onlyofficeCallbackHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var data OnlyOfficeCallback
	err = json.Unmarshal(body, &data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if data.Status == 2 || data.Status == 6 {
		docPath := r.URL.Query().Get("save")
		if docPath == "" {
			return http.StatusInternalServerError, errors.New("unable to get file save path")
		}

		if !d.user.Perm.Modify || !d.Check(docPath) {
			return http.StatusForbidden, nil
		}

		doc, err := http.Get(data.URL)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		defer doc.Body.Close()

		err = d.RunHook(func() error {
			_, writeErr := writeFile(d.user.Fs, docPath, doc.Body)
			if writeErr != nil {
				return writeErr
			}
			return nil
		}, "save", docPath, "", d.user)

		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	resp := map[string]int{
		"error": 0,
	}
	return renderJSON(w, r, resp)
})

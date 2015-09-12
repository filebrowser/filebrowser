package routing

import (
	"net/http"
	"strings"

	"github.com/hacdias/caddy-hugo/edit"
	"github.com/mholt/caddy/middleware"
)

const (
	mainURL     string = "/admin"
	contentURL  string = mainURL + "/content"
	dataURL     string = mainURL + "/data"
	editURL     string = mainURL + "/edit"
	newURL      string = mainURL + "/new"
	settingsURL string = mainURL + "/settings"
	staticURL   string = mainURL + "/static"
)

// Route the admin path
func Route(w http.ResponseWriter, r *http.Request) (int, error) {
	if middleware.Path(r.URL.Path).Matches(contentURL) {
		w.Write([]byte("Show Content Folder"))
	} else if middleware.Path(r.URL.Path).Matches(dataURL) {
		w.Write([]byte("Show Data Folder"))
	} else if middleware.Path(r.URL.Path).Matches(editURL) {
		return edit.Execute(w, r, strings.Replace(r.URL.Path, editURL+"/", "", 1))
	} else if middleware.Path(r.URL.Path).Matches(newURL) {
		w.Write([]byte("New Thing Page"))
	} else if middleware.Path(r.URL.Path).Matches(settingsURL) {
		w.Write([]byte("Settings Page"))
	} else if middleware.Path(r.URL.Path).Matches(staticURL) {
		w.Write([]byte("Static things management"))
	} else if r.URL.Path == mainURL || r.URL.Path == mainURL+"/" {
		w.Write([]byte("Dashboard"))
	} else {
		return 404, nil
	}

	return 200, nil
}

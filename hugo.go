//go:generate go-bindata -pkg assets -o assets/assets.go static/css/ templates/

package hugo

import (
	"net/http"
	"strings"

	"github.com/hacdias/caddy-hugo/assets"
	"github.com/hacdias/caddy-hugo/edit"
	"github.com/mholt/caddy/config/setup"
	"github.com/mholt/caddy/middleware"
	"github.com/spf13/hugo/commands"
)

const (
	mainURL     string = "/admin"
	contentURL  string = mainURL + "/content"
	browseURL   string = mainURL + "/browse"
	editURL     string = mainURL + "/edit"
	settingsURL string = mainURL + "/settings"
	assetsURL   string = mainURL + "/assets"
)

// Setup function
func Setup(c *setup.Controller) (middleware.Middleware, error) {
	commands.Execute()

	return func(next middleware.Handler) middleware.Handler {
		return &handler{Next: next}
	}, nil
}

type handler struct{ Next middleware.Handler }

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if urlMatch(r, "/admin") {
		return route(w, r)
	}

	return h.Next.ServeHTTP(w, r)
}

func route(w http.ResponseWriter, r *http.Request) (int, error) {
	if urlMatch(r, contentURL) {
		// content folder management
		w.Write([]byte("Show Content Folder"))
	} else if urlMatch(r, browseURL) {
		// browse files
		w.Write([]byte("Show Data Folder"))
	} else if urlMatch(r, editURL) {
		// edit file
		return edit.Execute(w, r, strings.Replace(r.URL.Path, editURL+"/", "", 1))
	} else if urlMatch(r, settingsURL) {
		// edit settings
		w.Write([]byte("Settings Page"))

	} else if urlMatch(r, assetsURL) {
		// assets like css, javascript and images
		fileName := strings.Replace(r.URL.Path, assetsURL, "static", 1)
		file, err := assets.Asset(fileName)

		if err != nil {
			return 404, nil
		}

		w.Write(file)
	} else if r.URL.Path == mainURL || r.URL.Path == mainURL+"/" {
		// dashboard
		w.Write([]byte("Dashboard"))
	} else {
		return 404, nil
	}

	return 200, nil
}

func urlMatch(r *http.Request, str string) bool {
	return middleware.Path(r.URL.Path).Matches(str)
}

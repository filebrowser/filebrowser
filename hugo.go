//go:generate go-bindata -pkg assets -o assets/assets.go static/css/ static/js/ templates/

package hugo

import (
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-hugo/assets"
	"github.com/hacdias/caddy-hugo/edit"
	"github.com/mholt/caddy/config/setup"
	"github.com/mholt/caddy/middleware"
	"github.com/spf13/hugo/commands"
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
	page := parseComponents(r)[1]

	if page == "assets" {
		filename := strings.Replace(r.URL.Path, assetsURL, "static", 1)
		file, err := assets.Asset(filename)

		if err != nil {
			return 404, nil
		}

		extension := filepath.Ext(filename)
		mime := mime.TypeByExtension(extension)

		header := w.Header()
		header.Set("Content-Type", mime)

		w.Write(file)
	} else if page == "content" {
		w.Write([]byte("Content Page"))
	} else if page == "browse" {
		w.Write([]byte("Show Data Folder"))
	} else if page == "edit" {
		return edit.Execute(w, r)
	} else if page == "settings" {
		w.Write([]byte("Settings Page"))
	} else {
		return 404, nil
	}

	return 200, nil
}

func parseComponents(r *http.Request) []string {
	//The URL that the user queried.
	path := r.URL.Path
	path = strings.TrimSpace(path)
	//Cut off the leading and trailing forward slashes, if they exist.
	//This cuts off the leading forward slash.
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	//This cuts off the trailing forward slash.
	if strings.HasSuffix(path, "/") {
		cutOffLastCharLen := len(path) - 1
		path = path[:cutOffLastCharLen]
	}
	//We need to isolate the individual components of the path.
	components := strings.Split(path, "/")
	return components
}

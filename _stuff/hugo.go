// Package hugo makes the bridge between the static website generator Hugo
// and the webserver Caddy, also providing an administrative user interface.
package hugo

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// Hugo contais the next middleware to be run and the configuration
// of the current one.
type Hugo struct {
	FileManager *filemanager.FileManager
	Next        httpserver.Handler
	Config      *Config
}

// ServeHTTP is the main function of the whole plugin that routes every single
// request to its function.
func (h Hugo) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	// Check if the current request if for this plugin
	if httpserver.Path(r.URL.Path).Matches(h.Config.Admin) {
		// If this request requires a raw file or a download, return the FileManager
		query := r.URL.Query()
		if val, ok := query["raw"]; ok && val[0] == "true" {
			return h.FileManager.ServeHTTP(w, r)
		}

		if val, ok := query["download"]; ok && val[0] == "true" {
			return h.FileManager.ServeHTTP(w, r)
		}

		// If the url matches exactly with /{admin}/settings/, redirect
		// to the page of the configuration file
		if r.URL.Path == h.Config.Admin+"/settings/" {
			var frontmatter string

			if _, err := os.Stat(h.Config.Path + "config.yaml"); err == nil {
				frontmatter = "yaml"
			}

			if _, err := os.Stat(h.Config.Path + "config.json"); err == nil {
				frontmatter = "json"
			}

			if _, err := os.Stat(h.Config.Path + "config.toml"); err == nil {
				frontmatter = "toml"
			}

			http.Redirect(w, r, h.Config.Admin+"/config."+frontmatter, http.StatusTemporaryRedirect)
			return 0, nil
		}

		filename := strings.Replace(r.URL.Path, c.Admin+"/edit/", "", 1)
		filename = c.Path + filename

		if strings.HasPrefix(r.URL.Path, h.Config.Admin+"/api/git/") && r.Method == http.MethodPost {
			return HandleGit(w, r, h.Config)
		}

		if h.ShouldHandle(r.URL) {
			// return editor
			return 0, nil
		}

		return h.FileManager.ServeHTTP(w, r)
	}

	return h.Next.ServeHTTP(w, r)
}

var extensions = []string{
	"md", "markdown", "mdown", "mmark",
	"asciidoc", "adoc", "ad",
	"rst",
	"html", "htm",
	"js",
	"toml", "yaml", "json",
}

// ShouldHandle checks if this extension should be handled by this plugin
func (h Hugo) ShouldHandle(url *url.URL) bool {
	extension := strings.TrimPrefix(filepath.Ext(url.Path), ".")

	for _, ext := range extensions {
		if ext == extension {
			return true
		}
	}

	return false
}

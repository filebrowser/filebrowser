//go:generate go get github.com/jteeuwen/go-bindata
//go:generate go install github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -nomemcopy -pkg hugo -prefix "assets" -o binary.go assets/...

// Package hugo makes the bridge between the static website generator Hugo
// and the webserver Caddy, also providing an administrative user interface.
package hugo

import (
	"bytes"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/hacdias/caddy-filemanager"
	"github.com/hacdias/caddy-filemanager/assets"
	"github.com/hacdias/caddy-filemanager/frontmatter"
	"github.com/hacdias/caddy-filemanager/handlers"
	"github.com/hacdias/caddy-filemanager/utils/variables"
	"github.com/hacdias/caddy-hugo/utils/commands"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/robfig/cron"
	"github.com/spf13/hugo/parser"
)

// Hugo is hugo
type Hugo struct {
	Next        httpserver.Handler
	Config      *Config
	FileManager *filemanager.FileManager
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
func (h Hugo) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	// If the site matches the baseURL
	if httpserver.Path(r.URL.Path).Matches(h.Config.BaseURL) {
		// Serve the hugo assets
		if httpserver.Path(r.URL.Path).Matches(h.Config.BaseURL + AssetsURL) {
			return serveAssets(w, r, h.Config)
		}

		// Serve the filemanager assets
		if httpserver.Path(r.URL.Path).Matches(h.Config.BaseURL + assets.BaseURL) {
			return h.FileManager.ServeHTTP(w, r)
		}

		// If the url matches exactly with /{admin}/settings/ serve that page
		// page variable isn't used here to avoid people using URLs like
		// "/{admin}/settings/something".
		if r.URL.Path == h.Config.BaseURL+"/settings/" || r.URL.Path == h.Config.BaseURL+"/settings" {
			var frontmatter string
			var err error

			if _, err = os.Stat(h.Config.Root + "config.yaml"); err == nil {
				frontmatter = "yaml"
			}

			if _, err = os.Stat(h.Config.Root + "config.json"); err == nil {
				frontmatter = "json"
			}

			if _, err = os.Stat(h.Config.Root + "config.toml"); err == nil {
				frontmatter = "toml"
			}

			http.Redirect(w, r, h.FileManager.Configs[0].AbsoluteURL()+"/config."+frontmatter, http.StatusTemporaryRedirect)
			return 0, nil
		}

		if r.Method == http.MethodPost && r.Header.Get("archetype") != "" {
			filename := r.Header.Get("Filename")
			archetype := r.Header.Get("archetype")

			if !strings.HasSuffix(filename, ".md") && !strings.HasSuffix(filename, ".markdown") {
				return h.FileManager.ServeHTTP(w, r)
			}

			filename = strings.Replace(r.URL.Path, h.Config.BaseURL+"/content/", "", 1) + filename
			filename = filepath.Clean(filename)

			args := []string{"new", filename, "--kind", archetype}

			if err := commands.Run(h.Config.Hugo, args, h.Config.Root); err != nil {
				return http.StatusInternalServerError, err
			}

			return http.StatusOK, nil
		}

		if canBeEdited(r.URL.Path) && r.Method == http.MethodPut {
			code, err := h.FileManager.ServeHTTP(w, r)

			if err != nil {
				return code, err
			}

			if r.Header.Get("Regenerate") == "true" {
				RunHugo(h.Config, false)
			}

			if r.Header.Get("Schedule") != "" {
				code, err = h.Schedule(w, r)
			}

			return code, err
		}

		return h.FileManager.ServeHTTP(w, r)
	}

	return h.Next.ServeHTTP(w, r)
}

// RunHugo is used to run the static website generator
func RunHugo(c *Config, force bool) {
	os.RemoveAll(c.Root + "public")

	// Prevent running if watching is enabled
	if b, pos := variables.StringInSlice("--watch", c.Args); b && !force {
		if len(c.Args) > pos && c.Args[pos+1] != "false" {
			return
		}

		if len(c.Args) == pos+1 {
			return
		}
	}

	if err := commands.Run(c.Hugo, c.Args, c.Root); err != nil {
		log.Println(err)
	}
}

// Schedule schedules a post to be published later
func (h Hugo) Schedule(w http.ResponseWriter, r *http.Request) (int, error) {
	t, err := time.Parse("2006-01-02T15:04", r.Header.Get("Schedule"))

	if err != nil {
		return http.StatusInternalServerError, err
	}

	scheduler := cron.New()
	scheduler.AddFunc(t.Format("05 04 15 02 01 *"), func() {
		filename := r.URL.Path
		filename = strings.Replace(filename, h.FileManager.Configs[0].WebDavURL, h.Config.Root, 1)
		filename = filepath.Clean(filename)

		raw, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Println(err)
			return
		}

		buffer := bytes.NewBuffer(raw)
		page, err := parser.ReadFrom(buffer)

		if err != nil {
			log.Println(err)
			return
		}

		content := strings.TrimSpace(string(page.Content()))
		front, err := frontmatter.Unmarshal(page.FrontMatter())

		if err != nil {
			log.Println(err)
			return
		}

		kind := reflect.TypeOf(front)

		if kind == reflect.TypeOf(map[interface{}]interface{}{}) {
			delete(front.(map[interface{}]interface{}), "draft")
			delete(front.(map[interface{}]interface{}), "Draft")
		} else {
			delete(front.(map[string]interface{}), "draft")
			delete(front.(map[string]interface{}), "Draft")
		}

		fm, err := handlers.ParseFrontMatter(front, h.FileManager.Configs[0].FrontMatter)

		if err != nil {
			log.Println(err)
			return
		}

		f := new(bytes.Buffer)
		f.Write(fm)
		f.Write([]byte(content))
		file := f.Bytes()

		if err = ioutil.WriteFile(filename, file, 0666); err != nil {
			return
		}

		RunHugo(h.Config, false)
	})
	scheduler.Start()

	return http.StatusOK, nil
}

func canBeEdited(name string) bool {
	extensions := [...]string{
		".md", ".markdown", ".mdown", ".mmark",
		".asciidoc", ".adoc", ".ad",
		".rst",
		".json", ".toml", ".yaml", ".csv", ".xml", ".rss", ".conf", ".ini",
		".tex", ".sty",
		".css", ".sass", ".scss",
		".js",
		".html",
		".txt", ".rtf",
		".sh", ".bash", ".ps1", ".bat", ".cmd",
		".php", ".pl", ".py",
		"Caddyfile",
		".c", ".cc", ".h", ".hh", ".cpp", ".hpp", ".f90",
		".f", ".bas", ".d", ".ada", ".nim", ".cr", ".java", ".cs", ".vala", ".vapi",
	}

	for _, extension := range extensions {
		if strings.HasSuffix(name, extension) {
			return true
		}
	}

	return false
}

// serveAssets provides the needed assets for the front-end
func serveAssets(w http.ResponseWriter, r *http.Request, c *Config) (int, error) {
	// gets the filename to be used with Assets function
	filename := strings.Replace(r.URL.Path, c.BaseURL+AssetsURL, "public", 1)
	file, err := Asset(filename)
	if err != nil {
		return http.StatusNotFound, nil
	}

	// Get the file extension and its mimetype
	extension := filepath.Ext(filename)
	mediatype := mime.TypeByExtension(extension)

	// Write the header with the Content-Type and write the file
	// content to the buffer
	w.Header().Set("Content-Type", mediatype)
	w.Write(file)
	return 200, nil
}

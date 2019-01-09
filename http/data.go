package http

import (
	"log"
	"net/http"
	"strconv"

	"github.com/filebrowser/filebrowser/v2/runner"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
)

type handleFunc func(w http.ResponseWriter, r *http.Request, d *data) (int, error)

type data struct {
	*runner.Runner
	settings *settings.Settings
	server   *settings.Server
	store    *storage.Storage
	user     *users.User
	raw      interface{}
}

// Check implements rules.Checker.
func (d *data) Check(path string) bool {
	for _, rule := range d.user.Rules {
		if rule.Matches(path) {
			return rule.Allow
		}
	}

	for _, rule := range d.settings.Rules {
		if rule.Matches(path) {
			return rule.Allow
		}
	}

	return true
}

func handle(fn handleFunc, prefix string, storage *storage.Storage, server *settings.Server) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		settings, err := storage.Settings.Get()
		if err != nil {
			log.Fatalln("ERROR: couldn't get settings")
			return
		}

		status, err := fn(w, r, &data{
			Runner:   &runner.Runner{Settings: settings},
			store:    storage,
			settings: settings,
			server:   server,
		})

		if status != 0 {
			txt := http.StatusText(status)
			http.Error(w, strconv.Itoa(status)+" "+txt, status)
		}

		if status >= 400 || err != nil {
			log.Printf("%s: %v %s %v", r.URL.Path, status, r.RemoteAddr, err)
		}
	})

	return http.StripPrefix(prefix, handler)
}

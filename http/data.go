package fbhttp

import (
	"log"
	"net/http"
	gopath "path"
	"strconv"

	"github.com/tomasen/realip"

	"github.com/filebrowser/filebrowser/v2/rules"
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

	// checkerPrefix is prepended to every path before evaluating rules. It is
	// set when the user's filesystem has been rebased onto a subdirectory (as
	// done for public shares), so that rules — which are relative to the user's
	// original scope — are still matched against the real path instead of the
	// rebased one. Empty for regular requests.
	checkerPrefix string
}

// Check implements rules.Checker.
func (d *data) Check(path string) bool {
	// When the filesystem has been rebased (e.g. a public share rooted at a
	// subdirectory), the incoming path is relative to that root. Resolve it
	// back to the user's original scope before matching rules, otherwise rules
	// targeting paths below the share root would be silently bypassed.
	if d.checkerPrefix != "" {
		path = gopath.Join(d.checkerPrefix, path)
	}

	if d.user.HideDotfiles && rules.MatchHidden(path) {
		return false
	}

	allow := true
	for _, rule := range d.settings.Rules {
		if rule.Matches(path) {
			allow = rule.Allow
		}
	}

	for _, rule := range d.user.Rules {
		if rule.Matches(path) {
			allow = rule.Allow
		}
	}

	return allow
}

func handle(fn handleFunc, prefix string, store *storage.Storage, server *settings.Server) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range globalHeaders {
			w.Header().Set(k, v)
		}

		settings, err := store.Settings.Get()
		if err != nil {
			log.Fatalf("ERROR: couldn't get settings: %v\n", err)
			return
		}

		status, err := fn(w, r, &data{
			Runner:   &runner.Runner{Enabled: server.EnableExec, Settings: settings},
			store:    store,
			settings: settings,
			server:   server,
		})

		if status >= 400 || err != nil {
			clientIP := realip.FromRequest(r)
			log.Printf("%s: %v %s %v", r.URL.Path, status, clientIP, err)
		}

		if status != 0 {
			txt := http.StatusText(status)
			if status == http.StatusBadRequest && err != nil {
				txt += " (" + err.Error() + ")"
			}
			http.Error(w, strconv.Itoa(status)+" "+txt, status)
			return
		}
	})

	return stripPrefix(prefix, handler)
}

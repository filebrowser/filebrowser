// Package filemanager provides middleware for managing files in a directory
// when directory path is requested instead of a specific file. Based on browse
// middleware.
package filemanager

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/webdav"

	. "github.com/hacdias/filemanager"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("filemanager", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

type plugin struct {
	Next    httpserver.Handler
	Configs []*config
}

type config struct {
	*FileManager
	baseURL string
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
func (f plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	for i := range f.Configs {
		// Checks if this Path should be handled by File Manager.
		if !httpserver.Path(r.URL.Path).Matches(f.Configs[i].baseURL) {
			continue
		}

		return f.Configs[i].ServeHTTP(w, r)
	}

	return f.Next.ServeHTTP(w, r)
}

// setup configures a new FileManager middleware instance.
func setup(c *caddy.Controller) error {
	configs, err := parse(c)
	if err != nil {
		return err
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return plugin{Configs: configs, Next: next}
	})

	return nil
}

func parse(c *caddy.Controller) ([]*config, error) {
	var (
		configs []*config
	)

	for c.Next() {
		baseURL := "/"
		baseScope := "."
		database := ""

		// Get the baseURL and baseScope
		args := c.RemainingArgs()

		if len(args) == 1 {
			baseURL = args[0]
		}

		if len(args) > 1 {
			baseScope = args[1]
		}

		for c.NextBlock() {
			switch c.Val() {
			case "database":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}

				database = c.Val()
			}
		}

		caddyConf := httpserver.GetConfig(c)

		// If there is no database path on the settings,
		// store one in .caddy/filemanager/name.db.
		if database == "" {
			path := filepath.Join(caddy.AssetsPath(), "filemanager")
			err := os.MkdirAll(path, 0700)
			if err != nil {
				return nil, err
			}

			// The name of the database is the hashed value of a string composed
			// by the host, address path and the baseurl of this File Manager
			// instance.
			hasher := sha256.New()
			hasher.Write([]byte(caddyConf.Addr.Host + caddyConf.Addr.Path + baseURL))
			sha := hex.EncodeToString(hasher.Sum(nil))
			database = filepath.Join(path, sha+".db")
		}

		fm, err := New(database, User{
			Username:      "admin",
			Password:      "admin",
			AllowCommands: true,
			AllowEdit:     true,
			AllowNew:      true,
			Commands:      []string{"git", "svn", "hg"},
			Rules: []*Rule{{
				Regex:  true,
				Allow:  false,
				Regexp: &Regexp{Raw: "\\/\\..+"},
			}},
			CSS:        "",
			FileSystem: webdav.Dir(baseScope),
		})

		if err != nil {
			return nil, err
		}

		m := &config{FileManager: fm}
		m.SetBaseURL(baseURL)
		m.SetPrefixURL(strings.TrimSuffix(caddyConf.Addr.Path, "/"))
		m.baseURL = strings.TrimSuffix(baseURL, "/")

		configs = append(configs, m)
	}

	return configs, nil
}

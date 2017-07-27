// Package filemanager provides middleware for managing files in a directory
// when directory path is requested instead of a specific file. Based on browse
// middleware.
package filemanager

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/hacdias/filemanager"
	"github.com/hacdias/fileutils"
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

		f.Configs[i].ServeHTTP(w, r)
		return 0, nil
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

		if len(args) >= 1 {
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

		path := filepath.Join(caddy.AssetsPath(), "filemanager")
		err := os.MkdirAll(path, 0700)
		if err != nil {
			return nil, err
		}

		// if there is a database path and it is not absolute,
		// it will be relative to Caddy folder.
		if !filepath.IsAbs(database) && database != "" {
			database = filepath.Join(path, database)
		}

		// If there is no database path on the settings,
		// store one in .caddy/filemanager/name.db.
		if database == "" {
			// The name of the database is the hashed value of a string composed
			// by the host, address path and the baseurl of this File Manager
			// instance.
			hasher := md5.New()
			hasher.Write([]byte(caddyConf.Addr.Host + caddyConf.Addr.Path + baseURL))
			sha := hex.EncodeToString(hasher.Sum(nil))
			database = filepath.Join(path, sha+".db")

			fmt.Println("[WARNING] A database is going to be created for your File Manager instace at " + database +
				". It is highly recommended that you set the 'database' option to '" + sha + ".db'\n")
		}

		fm, err := New(database, User{
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
			FileSystem: fileutils.Dir(baseScope),
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

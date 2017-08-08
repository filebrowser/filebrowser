package hugo

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hacdias/filemanager"
	"github.com/hacdias/filemanager/plugins"
	"github.com/hacdias/fileutils"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

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

func parse(c *caddy.Controller) ([]*filemanager.FileManager, error) {
	var (
		configs []*filemanager.FileManager
	)

	for c.Next() {
		// hugo [directory] [admin] {
		// 		database path
		// }
		directory := "."
		admin := "/admin"
		database := ""
		noAuth := false

		// Get the baseURL and baseScope
		args := c.RemainingArgs()

		if len(args) >= 1 {
			directory = args[0]
		}

		if len(args) > 1 {
			admin = args[1]
		}

		for c.NextBlock() {
			switch c.Val() {
			case "database":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}

				database = c.Val()
			case "no_auth":
				if !c.NextArg() {
					noAuth = true
					continue
				}

				var err error
				noAuth, err = strconv.ParseBool(c.Val())
				if err != nil {
					return nil, err
				}
			}
		}

		caddyConf := httpserver.GetConfig(c)

		path := filepath.Join(caddy.AssetsPath(), "hugo")
		err := os.MkdirAll(path, 0700)
		if err != nil {
			return nil, err
		}

		// if there is a database path and it is not absolute,
		// it will be relative to ".caddy" folder.
		if !filepath.IsAbs(database) && database != "" {
			database = filepath.Join(path, database)
		}

		// If there is no database path on the settings,
		// store one in .caddy/hugo/{name}.db.
		if database == "" {
			// The name of the database is the hashed value of a string composed
			// by the host, address path and the baseurl of this File Manager
			// instance.
			hasher := md5.New()
			hasher.Write([]byte(caddyConf.Addr.Host + caddyConf.Addr.Path + admin))
			sha := hex.EncodeToString(hasher.Sum(nil))
			database = filepath.Join(path, sha+".db")

			fmt.Println("[WARNING] A database is going to be created for your Hugo instace at " + database +
				". It is highly recommended that you set the 'database' option to '" + sha + ".db'\n")
		}

		m, err := filemanager.New(database, filemanager.User{
			Locale:        "en",
			AllowCommands: true,
			AllowEdit:     true,
			AllowNew:      true,
			Permissions:   map[string]bool{},
			Commands:      []string{"git", "svn", "hg"},
			Rules: []*filemanager.Rule{{
				Regex:  true,
				Allow:  false,
				Regexp: &filemanager.Regexp{Raw: "\\/\\..+"},
			}},
			CSS:        "",
			FileSystem: fileutils.Dir(directory),
		})

		if err != nil {
			return nil, err
		}

		// Initialize the default settings for Hugo.
		hugo := &plugins.Hugo{
			Root:        directory,
			Public:      filepath.Join(directory, "public"),
			Args:        []string{},
			CleanPublic: true,
		}

		// Try to find the Hugo executable path.
		if err = hugo.Find(); err != nil {
			return nil, err
		}

		// Attaches Hugo plugin to this file manager instance.
		err = m.ActivatePlugin("hugo", hugo)
		if err != nil {
			return nil, err
		}

		m.NoAuth = noAuth
		m.SetBaseURL(admin)
		m.SetPrefixURL(strings.TrimSuffix(caddyConf.Addr.Path, "/"))
		configs = append(configs, m)
	}

	return configs, nil
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
func (p plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	for i := range p.Configs {
		// Checks if this Path should be handled by File Manager.
		if !httpserver.Path(r.URL.Path).Matches(p.Configs[i].BaseURL) {
			continue
		}

		p.Configs[i].ServeHTTP(w, r)
		return 0, nil
	}

	return p.Next.ServeHTTP(w, r)
}

func init() {
	caddy.RegisterPlugin("hugo", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

type plugin struct {
	Next    httpserver.Handler
	Configs []*filemanager.FileManager
}

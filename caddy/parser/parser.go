package parser

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/asdine/storm"
	"github.com/hacdias/filemanager"
	"github.com/hacdias/filemanager/bolt"
	"github.com/hacdias/filemanager/staticgen"
	"github.com/hacdias/fileutils"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/spf13/viper"
)

var databases = map[string]*storm.DB{}

// Parse ...
func Parse(c *caddy.Controller, plugin string) ([]*filemanager.FileManager, error) {
	var (
		configs []*filemanager.FileManager
		err     error
	)

	for c.Next() {
		u := &filemanager.User{
			Locale:        "en",
			AllowCommands: true,
			AllowEdit:     true,
			AllowNew:      true,
			AllowPublish:  true,
			Commands:      []string{"git", "svn", "hg"},
			CSS:           "",
			Rules: []*filemanager.Rule{{
				Regex:  true,
				Allow:  false,
				Regexp: &filemanager.Regexp{Raw: "\\/\\..+"},
			}},
		}

		baseURL := "/"
		scope := "."
		database := ""
		noAuth := false

		if plugin != "" {
			baseURL = "/admin"
		}

		// Get the baseURL and scope
		args := c.RemainingArgs()

		if plugin == "" {
			if len(args) >= 1 {
				baseURL = args[0]
			}

			if len(args) > 1 {
				scope = args[1]
			}
		} else {
			if len(args) >= 1 {
				scope = args[0]
			}

			if len(args) > 1 {
				baseURL = args[1]
			}
		}

		for c.NextBlock() {
			switch c.Val() {
			case "database":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}

				database = c.Val()
			case "locale":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}

				u.Locale = c.Val()
			case "allow_commands":
				if !c.NextArg() {
					u.AllowCommands = true
					continue
				}

				u.AllowCommands, err = strconv.ParseBool(c.Val())
				if err != nil {
					return nil, err
				}
			case "allow_edit":
				if !c.NextArg() {
					u.AllowEdit = true
					continue
				}

				u.AllowEdit, err = strconv.ParseBool(c.Val())
				if err != nil {
					return nil, err
				}
			case "allow_new":
				if !c.NextArg() {
					u.AllowNew = true
					continue
				}

				u.AllowNew, err = strconv.ParseBool(c.Val())
				if err != nil {
					return nil, err
				}
			case "allow_publish":
				if !c.NextArg() {
					u.AllowPublish = true
					continue
				}

				u.AllowPublish, err = strconv.ParseBool(c.Val())
				if err != nil {
					return nil, err
				}
			case "commands":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}

				u.Commands = strings.Split(c.Val(), " ")
			case "css":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}

				file := c.Val()
				css, err := ioutil.ReadFile(file)
				if err != nil {
					return nil, err
				}

				u.CSS = string(css)
			case "no_auth":
				if !c.NextArg() {
					noAuth = true
					continue
				}

				noAuth, err = strconv.ParseBool(c.Val())
				if err != nil {
					return nil, err
				}
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

		u.Scope = scope
		u.FileSystem = fileutils.Dir(scope)

		var db *storm.DB
		if stored, ok := databases[database]; ok {
			db = stored
		} else {
			db, err = storm.Open(database)
			databases[database] = db
		}

		if err != nil {
			return nil, err
		}

		m := &filemanager.FileManager{
			NoAuth:      viper.GetBool("NoAuth"),
			BaseURL:     "",
			PrefixURL:   "",
			DefaultUser: u,
			Store: &filemanager.Store{
				Config: bolt.ConfigStore{DB: db},
				Users:  bolt.UsersStore{DB: db},
				Share:  bolt.ShareStore{DB: db},
			},
			NewFS: func(scope string) filemanager.FileSystem {
				return fileutils.Dir(scope)
			},
		}

		err = m.Setup()
		if err != nil {
			return nil, err
		}

		switch plugin {
		case "hugo":
			// Initialize the default settings for Hugo.
			hugo := &staticgen.Hugo{
				Root:        scope,
				Public:      filepath.Join(scope, "public"),
				Args:        []string{},
				CleanPublic: true,
			}

			// Attaches Hugo plugin to this file manager instance.
			err = m.Attach(hugo)
			if err != nil {
				return nil, err
			}
		case "jekyll":
			// Initialize the default settings for Jekyll.
			jekyll := &staticgen.Jekyll{
				Root:        scope,
				Public:      filepath.Join(scope, "_site"),
				Args:        []string{},
				CleanPublic: true,
			}

			// Attaches Hugo plugin to this file manager instance.
			err = m.Attach(jekyll)
			if err != nil {
				return nil, err
			}
		}

		if err != nil {
			return nil, err
		}

		m.NoAuth = noAuth
		m.SetBaseURL(baseURL)
		m.SetPrefixURL(strings.TrimSuffix(caddyConf.Addr.Path, "/"))

		configs = append(configs, m)
	}

	return configs, nil
}

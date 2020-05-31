package importer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/asdine/storm"
	toml "github.com/pelletier/go-toml"
	yaml "gopkg.in/yaml.v2"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
)

type oldDefs struct {
	Commands      []string `json:"commands" yaml:"commands" toml:"commands"`
	Scope         string   `json:"scope" yaml:"scope" toml:"scope"`
	ViewMode      string   `json:"viewMode" yaml:"viewMode" toml:"viewMode"`
	Locale        string   `json:"locale" yaml:"locale" toml:"locale"`
	AllowCommands bool     `json:"allowCommands" yaml:"allowCommands" toml:"allowCommands"`
	AllowEdit     bool     `json:"allowEdit" yaml:"allowEdit" toml:"allowEdit"`
	AllowNew      bool     `json:"allowNew" yaml:"allowNew" toml:"allowNew"`
}

type oldAuth struct {
	Method string `json:"method" yaml:"method" toml:"method"` // default none proxy
	Header string `json:"header" yaml:"header" toml:"header"`
}

type oldConf struct {
	Port      string  `json:"port" yaml:"port" toml:"port"`
	BaseURL   string  `json:"baseURL" yaml:"baseURL" toml:"baseURL"`
	Log       string  `json:"log" yaml:"log" toml:"log"`
	Address   string  `json:"address" yaml:"address" toml:"address"`
	Defaults  oldDefs `json:"defaults" yaml:"defaults" toml:"defaults"`
	ReCaptcha struct {
		Key    string `json:"key" yaml:"key" toml:"key"`
		Secret string `json:"secret" yaml:"secret" toml:"secret"`
		Host   string `json:"host" yaml:"host" toml:"host"`
	} `json:"recaptcha" yaml:"recaptcha" toml:"recaptcha"`
	Auth oldAuth `json:"auth" yaml:"auth" toml:"auth"`
}

var defaults = &oldConf{
	Port: "0",
	Log:  "stdout",
	Defaults: oldDefs{
		Commands:      []string{"git", "svn", "hg"},
		ViewMode:      string(users.MosaicViewMode),
		AllowCommands: true,
		AllowEdit:     true,
		AllowNew:      true,
		Locale:        "en",
	},
	Auth: oldAuth{
		Method: "default",
	},
}

func readConf(path string) (*oldConf, error) {
	cfg := &oldConf{}
	if path != "" {
		ext := filepath.Ext(path)

		fd, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer fd.Close()

		switch ext {
		case ".json":
			err = json.NewDecoder(fd).Decode(cfg)
		case ".toml":
			err = toml.NewDecoder(fd).Decode(cfg)
		case ".yaml", ".yml":
			err = yaml.NewDecoder(fd).Decode(cfg)
		default:
			return nil, errors.New("unsupported config extension " + ext)
		}

		if err != nil {
			return nil, err
		}
	} else {
		cfg = defaults
		path, err := filepath.Abs(".")
		if err != nil {
			return nil, err
		}
		cfg.Defaults.Scope = path
	}
	return cfg, nil
}

func importConf(db *storm.DB, path string, sto *storage.Storage) error {
	cfg, err := readConf(path)
	if err != nil {
		return err
	}

	commands := map[string][]string{}
	err = db.Get("config", "commands", &commands)
	if err != nil {
		return err
	}

	key := []byte{}
	err = db.Get("config", "key", &key)
	if err != nil {
		return err
	}

	s := &settings.Settings{
		Key:    key,
		Signup: false,
		Defaults: settings.UserDefaults{
			Scope:    cfg.Defaults.Scope,
			Commands: cfg.Defaults.Commands,
			ViewMode: users.ViewMode(cfg.Defaults.ViewMode),
			Locale:   cfg.Defaults.Locale,
			Perm: users.Permissions{
				Admin:    false,
				Execute:  cfg.Defaults.AllowCommands,
				Create:   cfg.Defaults.AllowNew,
				Rename:   cfg.Defaults.AllowEdit,
				Modify:   cfg.Defaults.AllowEdit,
				Delete:   cfg.Defaults.AllowEdit,
				Share:    true,
				Download: true,
			},
		},
	}

	server := &settings.Server{
		BaseURL: cfg.BaseURL,
		Port:    cfg.Port,
		Address: cfg.Address,
		Log:     cfg.Log,
	}

	var auther auth.Auther
	switch cfg.Auth.Method {
	case "proxy":
		auther = &auth.ProxyAuth{Header: cfg.Auth.Header}
		s.AuthMethod = auth.MethodProxyAuth
	case "none":
		auther = &auth.NoAuth{}
		s.AuthMethod = auth.MethodNoAuth
	default:
		auther = &auth.JSONAuth{
			ReCaptcha: &auth.ReCaptcha{
				Host:   cfg.ReCaptcha.Host,
				Key:    cfg.ReCaptcha.Key,
				Secret: cfg.ReCaptcha.Secret,
			},
		}
		s.AuthMethod = auth.MethodJSONAuth
	}

	err = sto.Auth.Save(auther)
	if err != nil {
		return err
	}

	err = sto.Settings.Save(s)
	if err != nil {
		return err
	}

	err = sto.Settings.SaveServer(server)
	if err != nil {
		return err
	}

	fmt.Println("Configuration successfully imported.")
	return nil
}

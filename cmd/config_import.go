package cmd

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
)

func init() {
	configCmd.AddCommand(configImportCmd)
}

type settingsFile struct {
	Settings *settings.Settings `json:"settings"`
	Server   *settings.Server   `json:"server"`
	Auther   interface{}        `json:"auther"`
}

var configImportCmd = &cobra.Command{
	Use:   "import <path>",
	Short: "Import a configuration file",
	Long: `Import a configuration file. This will replace all the existing
configuration. Can be used with or without unexisting databases.

If used with a nonexisting database, a key will be generated
automatically. Otherwise the key will be kept the same as in the
database.

The path must be for a json or yaml file.`,
	Args: jsonYamlArg,
	RunE: python(func(_ *cobra.Command, args []string, d *pythonData) error {
		var key []byte
		var err error
		if d.hadDB {
			settings, settingErr := d.store.Settings.Get()
			if settingErr != nil {
				return settingErr
			}
			key = settings.Key
		} else {
			key = generateKey()
		}

		file := settingsFile{}
		err = unmarshal(args[0], &file)
		if err != nil {
			return err
		}

		file.Settings.Key = key
		err = d.store.Settings.Save(file.Settings)
		if err != nil {
			return err
		}

		err = d.store.Settings.SaveServer(file.Server)
		if err != nil {
			return err
		}

		var rawAuther interface{}
		if filepath.Ext(args[0]) != ".json" {
			rawAuther = cleanUpInterfaceMap(file.Auther.(map[interface{}]interface{}))
		} else {
			rawAuther = file.Auther
		}

		var auther auth.Auther
		var autherErr error
		switch file.Settings.AuthMethod {
		case auth.MethodJSONAuth:
			var a interface{}
			a, autherErr = getAuther(auth.JSONAuth{}, rawAuther)
			auther = a.(*auth.JSONAuth)
		case auth.MethodNoAuth:
			var a interface{}
			a, autherErr = getAuther(auth.NoAuth{}, rawAuther)
			auther = a.(*auth.NoAuth)
		case auth.MethodProxyAuth:
			var a interface{}
			a, autherErr = getAuther(auth.ProxyAuth{}, rawAuther)
			auther = a.(*auth.ProxyAuth)
		case auth.MethodHookAuth:
			var a interface{}
			a, autherErr = getAuther(&auth.HookAuth{}, rawAuther)
			auther = a.(*auth.HookAuth)
		default:
			return errors.New("invalid auth method")
		}

		if autherErr != nil {
			return autherErr
		}

		err = d.store.Auth.Save(auther)
		if err != nil {
			return err
		}

		return printSettings(file.Server, file.Settings, auther)
	}, pythonConfig{allowNoDB: true}),
}

func getAuther(sample auth.Auther, data interface{}) (interface{}, error) {
	authType := reflect.TypeOf(sample)
	auther := reflect.New(authType).Interface()
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &auther)
	if err != nil {
		return nil, err
	}
	return auther, nil
}

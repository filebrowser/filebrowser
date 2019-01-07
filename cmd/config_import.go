package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(configImportCmd)
}

type settingsFile struct {
	Settings *settings.Settings `json:"settings"`
	Auther   interface{}        `json:"auther"`
}

var configImportCmd = &cobra.Command{
	Use: "import <filename>",
	Short: `Import a configuration file. This will replace all the existing
configuration. Can be used with or without unexisting databases.
If used with a nonexisting database, a key will be generated
automatically. Otherwise the key will be kept the same as in the
database.`,
	Args: cobra.ExactArgs(1),
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		var key []byte
		if d.hadDB {
			settings, err := d.store.Settings.Get()
			checkErr(err)
			key = settings.Key
		} else {
			key = generateRandomBytes(64)
		}

		fd, err := os.Open(args[0])
		checkErr(err)
		defer fd.Close()

		file := settingsFile{}
		err = json.NewDecoder(fd).Decode(&file)
		checkErr(err)

		file.Settings.Key = key

		err = d.store.Settings.Save(file.Settings)
		checkErr(err)

		var auther auth.Auther
		switch file.Settings.AuthMethod {
		case auth.MethodJSONAuth:
			auther = getAuther(auth.JSONAuth{}, file.Auther).(*auth.JSONAuth)
		case auth.MethodNoAuth:
			auther = getAuther(auth.NoAuth{}, file.Auther).(*auth.NoAuth)
		case auth.MethodProxyAuth:
			auther = getAuther(auth.ProxyAuth{}, file.Auther).(*auth.ProxyAuth)
		default:
			checkErr(errors.New("invalid auth method"))
		}

		err = d.store.Auth.Save(auther)
		checkErr(err)
		printSettings(file.Settings, auther)
	}, pythonConfig{allowNoDB: true}),
}

func getAuther(sample auth.Auther, data interface{}) interface{} {
	authType := reflect.TypeOf(sample)
	auther := reflect.New(authType).Interface()

	bytes, err := json.Marshal(data)
	checkErr(err)
	err = json.Unmarshal(bytes, &auther)

	return auther
}

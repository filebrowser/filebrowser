package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/settings"
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configInitCmd)
	addConfigFlags(configInitCmd)
	configInitCmd.MarkFlagRequired("scope")
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new database",
	Long: `Initialize a new database to use with File Browser. All of
this options can be changed in the future with the command
"filebrowser config set". The user related flags apply
to the defaults when creating new users and you don't
override the options.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(databasePath); err == nil {
			panic(errors.New(databasePath + " already exists"))
		}

		defaults := settings.UserDefaults{}
		getUserDefaults(cmd, &defaults, true)
		authMethod, auther := getAuthentication(cmd)

		db, err := storm.Open(databasePath)
		checkErr(err)
		defer db.Close()
		st := getStorage(db)
		s := &settings.Settings{
			Key:        generateRandomBytes(64), // 256 bit
			BaseURL:    mustGetString(cmd, "baseURL"),
			Log:        mustGetString(cmd, "log"),
			Signup:     mustGetBool(cmd, "signup"),
			Shell:      strings.Split(strings.TrimSpace(mustGetString(cmd, "shell")), " "),
			AuthMethod: authMethod,
			Defaults:   defaults,
			Server: settings.Server{
				Address: mustGetString(cmd, "address"),
				Port:    mustGetInt(cmd, "port"),
				TLSCert: mustGetString(cmd, "tls.cert"),
				TLSKey:  mustGetString(cmd, "tls.key"),
			},
			Branding: settings.Branding{
				Name:            mustGetString(cmd, "branding.name"),
				DisableExternal: mustGetBool(cmd, "branding.disableExternal"),
				Files:           mustGetString(cmd, "branding.files"),
			},
		}

		err = st.Settings.Save(s)
		checkErr(err)
		err = st.Auth.Save(auther)
		checkErr(err)

		fmt.Printf(`
Congratulations! You've set up your database to use with File Browser.
Now add your first user via 'filebrowser users new' and then you just
need to call the main command to boot up the server.
`)
		printSettings(s, auther)
	},
}

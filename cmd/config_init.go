package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/types"
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

		defaults := types.UserDefaults{}
		getUserDefaults(cmd, &defaults, true)
		authMethod, auther := getAuthentication(cmd)

		settings := &types.Settings{
			Key:        generateRandomBytes(64), // 256 bits
			BaseURL:    mustGetString(cmd, "baseURL"),
			Signup:     mustGetBool(cmd, "signup"),
			Shell:      strings.Split(strings.TrimSpace(mustGetString(cmd, "shell")), " "),
			Defaults:   defaults,
			AuthMethod: authMethod,
			Branding: types.Branding{
				Name:            mustGetString(cmd, "branding.name"),
				DisableExternal: mustGetBool(cmd, "branding.disableExternal"),
				Files:           mustGetString(cmd, "branding.files"),
			},
		}

		db, err := storm.Open(databasePath)
		checkErr(err)
		defer db.Close()

		saveConfig(db, settings, auther)

		fmt.Printf(`
Congratulations! You've set up your database to use with File Browser.
Now add your first user via 'filebrowser users new' and then you just
need to call the main command to boot up the server.
`)
		printSettings(settings, auther)
	},
}

func saveConfig(db *storm.DB, s *types.Settings, a types.Auther) {
	st := getStore(db)
	err := st.SaveSettings(s)
	checkErr(err)
	err = st.SaveAuther(a)
	checkErr(err)
}

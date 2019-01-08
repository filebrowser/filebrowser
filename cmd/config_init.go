package cmd

import (
	"fmt"
	"strings"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(configInitCmd)
	addConfigFlags(configInitCmd.Flags())
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
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		defaults := settings.UserDefaults{}
		getUserDefaults(cmd, &defaults, true)
		authMethod, auther := getAuthentication(cmd)

		s := &settings.Settings{
			Key:        generateRandomBytes(64), // 256 bit
			Signup:     mustGetBool(cmd, "signup"),
			Shell:      strings.Split(strings.TrimSpace(mustGetString(cmd, "shell")), " "),
			AuthMethod: authMethod,
			Defaults:   defaults,
			Branding: settings.Branding{
				Name:            mustGetString(cmd, "branding.name"),
				DisableExternal: mustGetBool(cmd, "branding.disableExternal"),
				Files:           mustGetString(cmd, "branding.files"),
			},
		}

		err := d.store.Settings.Save(s)
		checkErr(err)
		err = d.store.Auth.Save(auther)
		checkErr(err)

		fmt.Printf(`
Congratulations! You've set up your database to use with File Browser.
Now add your first user via 'filebrowser users new' and then you just
need to call the main command to boot up the server.
`)
		printSettings(s, auther)
	}, pythonConfig{noDB: true}),
}

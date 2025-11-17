package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/filebrowser/filebrowser/v2/settings"
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
'filebrowser config set'. The user related flags apply
to the defaults when creating new users and you don't
override the options.`,
	Args: cobra.NoArgs,
	RunE: python(func(cmd *cobra.Command, _ []string, d *pythonData) error {
		flags := cmd.Flags()

		// Initialize config
		s := &settings.Settings{Key: generateKey()}
		ser := &settings.Server{}

		// Fill config with options
		auther, err := getSettings(flags, s, ser, nil, true)
		if err != nil {
			return err
		}

		// Save updated config
		err = d.store.Settings.Save(s)
		if err != nil {
			return err
		}

		err = d.store.Settings.SaveServer(ser)
		if err != nil {
			return err
		}

		err = d.store.Auth.Save(auther)
		if err != nil {
			return err
		}

		fmt.Printf(`
Congratulations! You've set up your database to use with File Browser.
Now add your first user via 'filebrowser users add' and then you just
need to call the main command to boot up the server.
`)
		return printSettings(ser, s, auther)
	}, pythonConfig{expectsNoDatabase: true}),
}

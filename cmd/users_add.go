package cmd

import (
	"github.com/spf13/cobra"

	"github.com/filebrowser/filebrowser/v2/users"
)

func init() {
	usersCmd.AddCommand(usersAddCmd)
	addUserFlags(usersAddCmd.Flags())
}

var usersAddCmd = &cobra.Command{
	Use:   "add <username> <password>",
	Short: "Create a new user",
	Long:  `Create a new user and add it to the database.`,
	Args:  cobra.ExactArgs(2),
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		s, err := d.store.Settings.Get()
		checkErr(err)
		getUserDefaults(cmd.Flags(), &s.Defaults, false)

		password, err := users.HashAndValidatePwd(args[1], s.MinimumPasswordLength)
		checkErr(err)

		user := &users.User{
			Username:     args[0],
			Password:     password,
			LockPassword: mustGetBool(cmd.Flags(), "lockPassword"),
		}

		s.Defaults.Apply(user)

		servSettings, err := d.store.Settings.GetServer()
		checkErr(err)
		// since getUserDefaults() polluted s.Defaults.Scope
		// which makes the Scope not the one saved in the db
		// we need the right s.Defaults.Scope here
		s2, err := d.store.Settings.Get()
		checkErr(err)

		userHome, err := s2.MakeUserDir(user.Username, user.Scope, servSettings.Root)
		checkErr(err)
		user.Scope = userHome

		err = d.store.Users.Save(user)
		checkErr(err)
		printUsers([]*users.User{user})
	}, pythonConfig{}),
}

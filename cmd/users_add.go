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
	RunE: python(func(cmd *cobra.Command, args []string, d *pythonData) error {
		s, err := d.store.Settings.Get()
		if err != nil {
			return err
		}
		err = getUserDefaults(cmd.Flags(), &s.Defaults, false)
		if err != nil {
			return err
		}

		password, err := users.ValidateAndHashPwd(args[1], s.MinimumPasswordLength)
		if err != nil {
			return err
		}

		lockPassword, err := getBool(cmd.Flags(), "lockPassword")
		if err != nil {
			return err
		}

		user := &users.User{
			Username:     args[0],
			Password:     password,
			LockPassword: lockPassword,
		}

		s.Defaults.Apply(user)

		servSettings, err := d.store.Settings.GetServer()
		if err != nil {
			return err
		}
		// since getUserDefaults() polluted s.Defaults.Scope
		// which makes the Scope not the one saved in the db
		// we need the right s.Defaults.Scope here
		s2, err := d.store.Settings.Get()
		if err != nil {
			return err
		}

		userHome, err := s2.MakeUserDir(user.Username, user.Scope, servSettings.Root)
		if err != nil {
			return err
		}
		user.Scope = userHome

		err = d.store.Users.Save(user)
		if err != nil {
			return err
		}
		printUsers([]*users.User{user})
		return nil
	}, pythonConfig{}),
}

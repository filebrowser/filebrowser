package cmd

import (
	"github.com/spf13/cobra"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

func init() {
	usersCmd.AddCommand(usersUpdateCmd)

	usersUpdateCmd.Flags().StringP("password", "p", "", "new password")
	usersUpdateCmd.Flags().StringP("username", "u", "", "new username")
	addUserFlags(usersUpdateCmd.Flags())
}

var usersUpdateCmd = &cobra.Command{
	Use:   "update <id|username>",
	Short: "Updates an existing user",
	Long: `Updates an existing user. Set the flags for the
options you want to change.`,
	Args: cobra.ExactArgs(1),
	RunE: python(func(cmd *cobra.Command, args []string, d *pythonData) error {
		username, id := parseUsernameOrID(args[0])
		flags := cmd.Flags()
		password, err := getString(flags, "password")
		if err != nil {
			return err
		}
		newUsername, err := getString(flags, "username")
		if err != nil {
			return err
		}

		s, err := d.store.Settings.Get()
		if err != nil {
			return err
		}

		var (
			user *users.User
		)

		if id != 0 {
			user, err = d.store.Users.Get("", id)
		} else {
			user, err = d.store.Users.Get("", username)
		}

		if err != nil {
			return err
		}

		defaults := settings.UserDefaults{
			Scope:       user.Scope,
			Locale:      user.Locale,
			ViewMode:    user.ViewMode,
			SingleClick: user.SingleClick,
			Perm:        user.Perm,
			Sorting:     user.Sorting,
			Commands:    user.Commands,
		}
		err = getUserDefaults(flags, &defaults, false)
		if err != nil {
			return err
		}
		user.Scope = defaults.Scope
		user.Locale = defaults.Locale
		user.ViewMode = defaults.ViewMode
		user.SingleClick = defaults.SingleClick
		user.Perm = defaults.Perm
		user.Commands = defaults.Commands
		user.Sorting = defaults.Sorting
		user.LockPassword, err = getBool(flags, "lockPassword")
		if err != nil {
			return err
		}

		if newUsername != "" {
			user.Username = newUsername
		}

		if password != "" {
			user.Password, err = users.ValidateAndHashPwd(password, s.MinimumPasswordLength)
			if err != nil {
				return err
			}
		}

		err = d.store.Users.Update(user)
		if err != nil {
			return err
		}
		printUsers([]*users.User{user})
		return nil
	}, pythonConfig{}),
}

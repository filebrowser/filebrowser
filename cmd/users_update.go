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
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		username, id := parseUsernameOrID(args[0])
		flags := cmd.Flags()
		password := mustGetString(flags, "password")
		newUsername := mustGetString(flags, "username")

		var (
			err  error
			user *users.User
		)

		if id != 0 {
			user, err = d.store.Users.Get("", id)
		} else {
			user, err = d.store.Users.Get("", username)
		}

		checkErr(err)

		defaults := settings.UserDefaults{
			Scope:    user.Scope,
			Locale:   user.Locale,
			ViewMode: user.ViewMode,
			Perm:     user.Perm,
			Sorting:  user.Sorting,
			Commands: user.Commands,
		}
		getUserDefaults(flags, &defaults, false)
		user.Scope = defaults.Scope
		user.Locale = defaults.Locale
		user.ViewMode = defaults.ViewMode
		user.Perm = defaults.Perm
		user.Commands = defaults.Commands
		user.Sorting = defaults.Sorting
		user.LockPassword = mustGetBool(flags, "lockPassword")

		if newUsername != "" {
			user.Username = newUsername
		}

		if password != "" {
			user.Password, err = users.HashPwd(password)
			checkErr(err)
		}

		err = d.store.Users.Update(user)
		checkErr(err)
		printUsers([]*users.User{user})
	}, pythonConfig{}),
}

package cmd

import (
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
)

func init() {
	usersCmd.AddCommand(usersUpdateCmd)

	usersUpdateCmd.Flags().StringP("password", "p", "", "new password")
	usersUpdateCmd.Flags().StringP("username", "u", "", "new username")
	addUserFlags(usersUpdateCmd)
}

var usersUpdateCmd = &cobra.Command{
	Use:   "update <id|username>",
	Short: "Updates an existing user",
	Long: `Updates an existing user. Set the flags for the
options you want to change.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()
		st := getStorage(db)

		set, err := st.Settings.Get()
		checkErr(err)

		username, id := parseUsernameOrID(args[0])
		password := mustGetString(cmd, "password")
		newUsername := mustGetString(cmd, "username")

		var user *users.User

		if id != 0 {
			user, err = st.Users.Get(set.Scope, id)
		} else {
			user, err = st.Users.Get(set.Scope, username)
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
		getUserDefaults(cmd, &defaults, false)
		user.Scope = defaults.Scope
		user.Locale = defaults.Locale
		user.ViewMode = defaults.ViewMode
		user.Perm = defaults.Perm
		user.Commands = defaults.Commands
		user.Sorting = defaults.Sorting
		user.LockPassword = mustGetBool(cmd, "lockPassword")

		if newUsername != "" {
			user.Username = newUsername
		}

		if password != "" {
			user.Password, err = users.HashPwd(password)
			checkErr(err)
		}

		err = st.Users.Update(user)
		checkErr(err)
		printUsers([]*users.User{user})
	},
}

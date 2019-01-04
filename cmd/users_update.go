package cmd

import (
	"github.com/filebrowser/filebrowser/lib"
	"github.com/spf13/cobra"
)

func init() {
	usersCmd.AddCommand(usersUpdateCmd)

	usersUpdateCmd.Flags().UintP("id", "i", 0, "id of the user")
	usersUpdateCmd.Flags().StringP("username", "u", "", "user to change or new username if flag 'id' is set")
	usersUpdateCmd.Flags().StringP("password", "p", "", "new password")
	addUserFlags(usersUpdateCmd)
}

var usersUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates an existing user",
	Long: `Updates an existing user. Set the flags for the
options you want to change.`,
	Args: usernameOrIDRequired,
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()
		st := getFileBrowser(db)

		id, _ := cmd.Flags().GetUint("id")
		username := mustGetString(cmd, "username")
		password := mustGetString(cmd, "password")

		var user *lib.User
		var err error

		if id != 0 {
			user, err = st.GetUser(id)
		} else {
			user, err = st.GetUser(username)
		}

		checkErr(err)

		defaults := lib.UserDefaults{
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

		if user.Username != username && username != "" {
			user.Username = username
		}

		if password != "" {
			user.Password, err = lib.HashPwd(password)
			checkErr(err)
		}

		err = st.UpdateUser(user)
		checkErr(err)
		printUsers([]*lib.User{user})
	},
}

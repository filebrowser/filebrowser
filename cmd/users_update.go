package cmd

import (
	storage "github.com/filebrowser/filebrowser/bolt"
	"github.com/filebrowser/filebrowser/types"
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
	Args:  usernameOrIDRequired,
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()
		st := storage.UsersStore{DB: db}

		id, _ := cmd.Flags().GetUint("id")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		var user *types.User
		var err error

		if id != 0 {
			user, err = st.Get(id)
		} else {
			user, err = st.GetByUsername(username)
		}

		checkErr(err)

		defaults := types.UserDefaults{
			Scope:    user.Scope,
			Locale:   user.Locale,
			ViewMode: user.ViewMode,
			Perm:     user.Perm,
		}
		getUserDefaults(cmd, &defaults, false)
		user.Scope = defaults.Scope
		user.Locale = defaults.Locale
		user.ViewMode = defaults.ViewMode
		user.Perm = defaults.Perm

		if user.Username != username && username != "" {
			user.Username = username
		}

		if password != "" {
			user.Password, err = types.HashPwd(password)
			checkErr(err)
		}

		err = st.Update(user)
		checkErr(err)
		printUsers([]*types.User{user})
	},
}

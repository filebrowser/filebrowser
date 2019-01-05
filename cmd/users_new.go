package cmd

import (
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
)

func init() {
	usersCmd.AddCommand(usersNewCmd)

	addUserFlags(usersNewCmd)
	usersNewCmd.Flags().StringP("username", "u", "", "new users's username")
	usersNewCmd.Flags().StringP("password", "p", "", "new user's password")
	usersNewCmd.MarkFlagRequired("username")
	usersNewCmd.MarkFlagRequired("password")
}

var usersNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new user",
	Long:  `Create a new user and add it to the database.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()
		st := getStorage(db)

		s, err := st.Settings.Get()
		checkErr(err)
		getUserDefaults(cmd, &s.Defaults, false)

		password, _ := cmd.Flags().GetString("password")
		password, err = users.HashPwd(password)
		checkErr(err)

		user := &users.User{
			Username:     mustGetString(cmd, "username"),
			Password:     password,
			LockPassword: mustGetBool(cmd, "lockPassword"),
		}

		s.Defaults.Apply(user)
		err = st.Users.Save(user)
		checkErr(err)
		printUsers([]*users.User{user})
	},
}

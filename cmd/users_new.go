package cmd

import (
	"github.com/filebrowser/filebrowser/types"
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
		st := getStore(db)

		settings, err := st.GetSettings()
		checkErr(err)
		getUserDefaults(cmd, &settings.Defaults, false)

		password, _ := cmd.Flags().GetString("password")
		password, err = types.HashPwd(password)
		checkErr(err)

		user := &types.User{
			Username:     mustGetString(cmd, "username"),
			Password:     password,
			LockPassword: mustGetBool(cmd, "lockPassword"),
		}

		user.ApplyDefaults(settings.Defaults)

		err = st.SaveUser(user)
		checkErr(err)
		printUsers([]*types.User{user})
	},
}

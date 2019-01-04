package cmd

import (
	"github.com/filebrowser/filebrowser/lib"
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
		st := getFileBrowser(db)

		settings := st.GetSettings()
		getUserDefaults(cmd, &settings.Defaults, false)

		password, _ := cmd.Flags().GetString("password")
		password, err := lib.HashPwd(password)
		checkErr(err)

		user := &lib.User{
			Username:     mustGetString(cmd, "username"),
			Password:     password,
			LockPassword: mustGetBool(cmd, "lockPassword"),
		}

		st.ApplyDefaults(user)
		err = st.SaveUser(user)
		checkErr(err)
		printUsers([]*lib.User{user})
	},
}

package cmd

import (
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
)

func init() {
	usersCmd.AddCommand(usersAddCmd)
	addUserFlags(usersAddCmd)
}

var usersAddCmd = &cobra.Command{
	Use:   "add <username> <password>",
	Short: "Create a new user",
	Long:  `Create a new user and add it to the database.`,
	Args:  cobra.ExactArgs(2),
	Run: python(func(cmd *cobra.Command, args []string, st *storage.Storage) {
		s, err := st.Settings.Get()
		checkErr(err)
		getUserDefaults(cmd, &s.Defaults, false)

		password, err := users.HashPwd(args[1])
		checkErr(err)

		user := &users.User{
			Username:     args[0],
			Password:     password,
			LockPassword: mustGetBool(cmd, "lockPassword"),
		}

		s.Defaults.Apply(user)
		err = st.Users.Save(user)
		checkErr(err)
		printUsers([]*users.User{user})
	}, pythonConfig{}),
}

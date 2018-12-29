package cmd

import (
	storage "github.com/filebrowser/filebrowser/bolt"
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
		ust := storage.UsersStore{DB: db}
		cst := storage.ConfigStore{DB: db}

		settings, err := cst.GetSettings()
		checkErr(err)
		getUserDefaults(cmd, &settings.Defaults, false)

		password, _ := cmd.Flags().GetString("password")
		password, err = types.HashPwd(password)
		checkErr(err)

		user := &types.User{
			Username:     mustGetString(cmd, "username"),
			Password:     password,
			LockPassword: mustGetBool(cmd, "lockPassword"),
			Scope:        settings.Defaults.Scope,
			Locale:       settings.Defaults.Locale,
			ViewMode:     settings.Defaults.ViewMode,
			Perm:         settings.Defaults.Perm,
		}

		err = ust.Save(user)
		checkErr(err)
		printUsers([]*types.User{user})
	},
}

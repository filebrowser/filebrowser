package cmd

import (
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
)

func init() {
	usersCmd.AddCommand(usersFindCmd)
	usersCmd.AddCommand(usersLsCmd)
	usersFindCmd.Flags().StringP("username", "u", "", "username to find")
	usersFindCmd.Flags().UintP("id", "i", 0, "id to find")
}

var usersFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find a user by username or id",
	Long:  `Find a user by username or id. If no flag is set, all users will be printed.`,
	Args:  cobra.NoArgs,
	Run:   findUsers,
}

var usersLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all users.",
	Args:  cobra.NoArgs,
	Run:   findUsers,
}

var findUsers = func(cmd *cobra.Command, args []string) {
	db := getDB()
	defer db.Close()
	st := getStorage(db)

	username, _ := cmd.Flags().GetString("username")
	id, _ := cmd.Flags().GetUint("id")

	var err error
	var list []*users.User
	var user *users.User

	if username != "" {
		user, err = st.Users.Get(username)
	} else if id != 0 {
		user, err = st.Users.Get(id)
	} else {
		list, err = st.Users.Gets()
	}

	checkErr(err)

	if user != nil {
		list = []*users.User{user}
	}

	printUsers(list)
}

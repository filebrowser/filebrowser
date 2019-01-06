package cmd

import (
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
)

func init() {
	usersCmd.AddCommand(usersFindCmd)
	usersCmd.AddCommand(usersLsCmd)
}

var usersFindCmd = &cobra.Command{
	Use:   "find <id|username>",
	Short: "Find a user by username or id",
	Long:  `Find a user by username or id. If no flag is set, all users will be printed.`,
	Args:  cobra.ExactArgs(1),
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

	var (
		list []*users.User
		user *users.User
		err  error
	)

	if len(args) == 1 {
		username, id := parseUsernameOrID(args[0])
		if username != "" {
			user, err = st.Users.Get("", username)
		} else {
			user, err = st.Users.Get("", id)
		}

		list = []*users.User{user}
	} else {
		list, err = st.Users.Gets("")
	}

	checkErr(err)
	printUsers(list)
}

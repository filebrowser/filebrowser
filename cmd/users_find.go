package cmd

import (
	"github.com/spf13/cobra"

	"github.com/filebrowser/filebrowser/v2/users"
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
	RunE:  findUsers,
}

var usersLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all users.",
	Args:  cobra.NoArgs,
	RunE:  findUsers,
}

var findUsers = withStore(func(_ *cobra.Command, args []string, st *store) error {
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

	if err != nil {
		return err
	}
	printUsers(list)
	return nil
}, storeOptions{})

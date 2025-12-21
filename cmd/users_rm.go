package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	usersCmd.AddCommand(usersRmCmd)
}

var usersRmCmd = &cobra.Command{
	Use:   "rm <id|username>",
	Short: "Delete a user by username or id",
	Long:  `Delete a user by username or id`,
	Args:  cobra.ExactArgs(1),
	RunE: withStore(func(_ *cobra.Command, args []string, st *store) error {
		username, id := parseUsernameOrID(args[0])
		var err error

		if username != "" {
			err = st.Users.Delete(username)
		} else {
			err = st.Users.Delete(id)
		}

		if err != nil {
			return err
		}
		fmt.Println("user deleted successfully")
		return nil
	}, storeOptions{}),
}

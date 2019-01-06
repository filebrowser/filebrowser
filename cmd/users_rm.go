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
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()
		st := getStorage(db)

		username, id := parseUsernameOrID(args[0])
		var err error

		if username != "" {
			err = st.Users.Delete(username)
		} else {
			err = st.Users.Delete(id)
		}

		checkErr(err)
		fmt.Println("user deleted successfully")
	},
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	usersCmd.AddCommand(usersRmCmd)
	usersRmCmd.Flags().StringP("username", "u", "", "username to delete")
	usersRmCmd.Flags().UintP("id", "i", 0, "id to delete")
}

var usersRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Delete a user by username or id",
	Long:  `Delete a user by username or id`,
	Args:  usernameOrIDRequired,
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()
		st := getStore(db)

		username, _ := cmd.Flags().GetString("username")
		id, _ := cmd.Flags().GetUint("id")

		var err error

		if username != "" {
			err = st.Users.DeleteByUsername(username)
		} else {
			err = st.Users.Delete(id)
		}

		checkErr(err)
		fmt.Println("user deleted successfully")
	},
}

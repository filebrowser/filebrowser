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
	RunE: python(func(_ *cobra.Command, args []string, d *pythonData) error {
		username, id := parseUsernameOrID(args[0])
		var err error

		if username != "" {
			err = d.store.Users.Delete(username)
		} else {
			err = d.store.Users.Delete(id)
		}

		if err != nil {
			return err
		}
		fmt.Println("user deleted successfully")
		return nil
	}, pythonConfig{}),
}

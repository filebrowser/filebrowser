package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
)

func init() {
	usersCmd.AddCommand(usersImportCmd)
	usersImportCmd.Flags().Bool("overwrite", false, "overwrite existing users,")
}

var usersImportCmd = &cobra.Command{
	Use:   "import <filename>",
	Short: "Import users from a file.",
	Args:  cobra.ExactArgs(1),
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		fd, err := os.Open(args[0])
		checkErr(err)
		defer fd.Close()

		list := []users.User{}
		err = json.NewDecoder(fd).Decode(&list)
		checkErr(err)

		overwrite := mustGetBool(cmd, "overwrite")

		for _, user := range list {
			// TODO: check for ID/Username conflicts too.
			_, err := d.store.Users.Get("", user.ID)
			if err == nil && !overwrite {
				checkErr(errors.New("user " + strconv.Itoa(int(user.ID)) + " is already registred"))
			}

			err = d.store.Users.Save(&user)
			checkErr(err)
		}

	}, pythonConfig{}),
}

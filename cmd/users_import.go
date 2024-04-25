package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/filebrowser/filebrowser/v2/users"
)

func init() {
	usersCmd.AddCommand(usersImportCmd)
	usersImportCmd.Flags().Bool("overwrite", false, "overwrite users with the same id/username combo")
	usersImportCmd.Flags().Bool("replace", false, "replace the entire user base")
}

var usersImportCmd = &cobra.Command{
	Use:   "import <path>",
	Short: "Import users from a file",
	Long: `Import users from a file. The path must be for a json or yaml
file. You can use this command to import new users to your
installation. For that, just don't place their ID on the files
list or set it to 0.`,
	Args: jsonYamlArg,
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		fd, err := os.Open(args[0])
		checkErr(err)
		defer fd.Close()

		list := []*users.User{}
		err = unmarshal(args[0], &list)
		checkErr(err)

		for _, user := range list {
			err = user.Clean("")
			checkErr(err)
		}

		if mustGetBool(cmd.Flags(), "replace") {
			oldUsers, err := d.store.Users.Gets("")
			checkErr(err)

			err = marshal("users.backup.json", list)
			checkErr(err)

			for _, user := range oldUsers {
				err = d.store.Users.Delete(user.ID)
				checkErr(err)
			}
		}

		overwrite := mustGetBool(cmd.Flags(), "overwrite")

		for _, user := range list {
			onDB, err := d.store.Users.Get("", user.ID)

			// User exists in DB.
			if err == nil {
				if !overwrite {
					checkErr(errors.New("user " + strconv.Itoa(int(user.ID)) + " is already registered"))
				}

				// If the usernames mismatch, check if there is another one in the DB
				// with the new username. If there is, print an error and cancel the
				// operation
				if user.Username != onDB.Username {
					if conflictuous, err := d.store.Users.Get("", user.Username); err == nil { //nolint:govet
						checkErr(usernameConflictError(user.Username, conflictuous.ID, user.ID))
					}
				}
			} else {
				// If it doesn't exist, set the ID to 0 to automatically get a new
				// one that make sense in this DB.
				user.ID = 0
			}

			err = d.store.Users.Save(user)
			checkErr(err)
		}
	}, pythonConfig{}),
}

func usernameConflictError(username string, originalID, newID uint) error {
	return fmt.Errorf(`can't import user with ID %d and username "%s" because the username is already registered with the user %d`,
		newID, username, originalID)
}

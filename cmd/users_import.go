package cmd

import (
	"errors"
	"os"
	"strconv"

	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
)

func init() {
	usersCmd.AddCommand(usersImportCmd)
	usersImportCmd.Flags().Bool("overwrite", false, "overwrite users with the same id/username combo")
	usersImportCmd.Flags().Bool("replace", false, "replace the entire user base")
}

var usersImportCmd = &cobra.Command{
	Use:   "import <filename>",
	Short: "Import users from a file.",
	Args:  jsonYamlArg,
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

		if mustGetBool(cmd, "replace") {
			oldUsers, err := d.store.Users.Gets("")
			checkErr(err)

			err = marshal("users.backup.json", list)
			checkErr(err)

			for _, user := range oldUsers {
				err = d.store.Users.Delete(user.ID)
				checkErr(err)
			}
		}

		overwrite := mustGetBool(cmd, "overwrite")

		for _, user := range list {
			onDB, err := d.store.Users.Get("", user.ID)

			// User exists in DB.
			if err == nil {
				if !overwrite {
					checkErr(errors.New("user " + strconv.Itoa(int(user.ID)) + " is already registred"))
				}

				// If the usernames mismatch, check if there is another one in the DB
				// with the new username. If there is, print an error and cancel the
				// operation
				if user.Username != onDB.Username {
					conflictuous, err := d.store.Users.Get("", user.Username)
					if err == nil {
						checkErr(usernameConflictError(user.Username, conflictuous.ID, user.ID))
					}
				}
			}

			err = d.store.Users.Save(user)
			checkErr(err)
		}
	}, pythonConfig{}),
}

func usernameConflictError(username string, original, new uint) error {
	return errors.New("can't import user with ID " + strconv.Itoa(int(new)) + " and username \"" + username + "\" because the username is already registred with the user " + strconv.Itoa(int(original)))
}

package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	usersCmd.AddCommand(usersExportCmd)
}

var usersExportCmd = &cobra.Command{
	Use:   "export <path>",
	Short: "Export all users to a file.",
	Long: `Export all users to a json or yaml file. Please indicate the
path to the file where you want to write the users.`,
	Args: jsonYamlArg,
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		list, err := d.store.Users.Gets("")
		checkErr(err)

		err = marshal(args[0], list)
		checkErr(err)
	}, pythonConfig{}),
}

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
	RunE: python(func(_ *cobra.Command, args []string, d *pythonData) error {
		list, err := d.store.Users.Gets("")
		if err != nil {
			return err
		}

		err = marshal(args[0], list)
		if err != nil {
			return err
		}
		return nil
	}, pythonConfig{}),
}

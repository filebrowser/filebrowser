package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	usersCmd.AddCommand(usersExportCmd)
}

var usersExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all users.",
	Args:  cobra.NoArgs,
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		list, err := d.store.Users.Gets("")
		checkErr(err)

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "    ")
		encoder.Encode(list)

	}, pythonConfig{}),
}

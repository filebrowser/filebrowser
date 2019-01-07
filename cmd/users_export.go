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
	Use:   "export <filename>",
	Short: "Export all users.",
	Args:  cobra.ExactArgs(1),
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		list, err := d.store.Users.Gets("")
		checkErr(err)

		fd, err := os.Create(args[0])
		checkErr(err)
		defer fd.Close()

		encoder := json.NewEncoder(fd)
		encoder.SetIndent("", "    ")
		encoder.Encode(list)

	}, pythonConfig{}),
}

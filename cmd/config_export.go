package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(configExportCmd)
}

var configExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the config.",
	Args:  cobra.NoArgs,
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		settings, err := d.store.Settings.Get()
		checkErr(err)

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "    ")
		encoder.Encode(settings)

	}, pythonConfig{}),
}

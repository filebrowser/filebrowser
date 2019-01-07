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
	Use:   "export <filename>",
	Short: "Export the configuration to a file.",
	Args:  cobra.ExactArgs(1),
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		settings, err := d.store.Settings.Get()
		checkErr(err)

		auther, err := d.store.Auth.Get(settings.AuthMethod)
		checkErr(err)

		fd, err := os.Create(args[0])
		checkErr(err)
		defer fd.Close()

		encoder := json.NewEncoder(fd)
		encoder.SetIndent("", "    ")
		encoder.Encode(&settingsFile{
			Settings: settings,
			Auther:   auther,
		})

	}, pythonConfig{}),
}

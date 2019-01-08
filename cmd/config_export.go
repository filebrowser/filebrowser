package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(configExportCmd)
}

var configExportCmd = &cobra.Command{
	Use:   "export <filename>",
	Short: "Export the configuration to a file.",
	Args:  jsonYamlArg,
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		settings, err := d.store.Settings.Get()
		checkErr(err)

		auther, err := d.store.Auth.Get(settings.AuthMethod)
		checkErr(err)

		data := &settingsFile{
			Settings: settings,
			Auther:   auther,
		}

		err = marshal(args[0], data)
		checkErr(err)
	}, pythonConfig{}),
}

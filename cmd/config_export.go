package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(configExportCmd)
}

var configExportCmd = &cobra.Command{
	Use:   "export <path>",
	Short: "Export the configuration to a file",
	Long: `Export the configuration to a file. The path must be for a
json or yaml file. This exported configuration can be changed,
and imported again with 'config import' command.`,
	Args: jsonYamlArg,
	Run: python(func(_ *cobra.Command, args []string, d pythonData) {
		settings, err := d.store.Settings.Get()
		checkErr(err)

		server, err := d.store.Settings.GetServer()
		checkErr(err)

		auther, err := d.store.Auth.Get(settings.AuthMethod)
		checkErr(err)

		data := &settingsFile{
			Settings: settings,
			Auther:   auther,
			Server:   server,
		}

		err = marshal(args[0], data)
		checkErr(err)
	}, pythonConfig{}),
}

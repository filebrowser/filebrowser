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
	RunE: withStore(func(_ *cobra.Command, args []string, st *store) error {
		settings, err := st.Settings.Get()
		if err != nil {
			return err
		}

		server, err := st.Settings.GetServer()
		if err != nil {
			return err
		}

		auther, err := st.Auth.Get(settings.AuthMethod)
		if err != nil {
			return err
		}

		data := &settingsFile{
			Settings: settings,
			Auther:   auther,
			Server:   server,
		}

		err = marshal(args[0], data)
		if err != nil {
			return err
		}
		return nil
	}, storeOptions{}),
}

package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(configSetCmd)
	addConfigFlags(configSetCmd.Flags())
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Updates the configuration",
	Long: `Updates the configuration. Set the flags for the options
you want to change. Other options will remain unchanged.`,
	Args: cobra.NoArgs,
	RunE: python(func(cmd *cobra.Command, _ []string, d *pythonData) error {
		flags := cmd.Flags()

		// Read existing config
		set, err := d.store.Settings.Get()
		if err != nil {
			return err
		}

		ser, err := d.store.Settings.GetServer()
		if err != nil {
			return err
		}

		auther, err := d.store.Auth.Get(set.AuthMethod)
		if err != nil {
			return err
		}

		// Get updated config
		auther, err = getSettings(flags, set, ser, auther, false)
		if err != nil {
			return err
		}

		// Save updated config
		err = d.store.Auth.Save(auther)
		if err != nil {
			return err
		}

		err = d.store.Settings.Save(set)
		if err != nil {
			return err
		}

		err = d.store.Settings.SaveServer(ser)
		if err != nil {
			return err
		}

		return printSettings(ser, set, auther)
	}, pythonConfig{}),
}

package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(configCatCmd)
}

var configCatCmd = &cobra.Command{
	Use:   "cat",
	Short: "Prints the configuration",
	Long:  `Prints the configuration.`,
	Args:  cobra.NoArgs,
	RunE: python(func(_ *cobra.Command, _ []string, d *pythonData) error {
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
		return printSettings(ser, set, auther)
	}, pythonConfig{}),
}

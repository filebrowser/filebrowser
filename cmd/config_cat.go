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
	Run: python(func(_ *cobra.Command, _ []string, d pythonData) {
		set, err := d.store.Settings.Get()
		checkErr(err)
		ser, err := d.store.Settings.GetServer()
		checkErr(err)
		auther, err := d.store.Auth.Get(set.AuthMethod)
		checkErr(err)
		printSettings(ser, set, auther)
	}, pythonConfig{}),
}

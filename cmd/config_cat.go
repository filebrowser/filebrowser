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
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		s, err := d.store.Settings.Get()
		checkErr(err)
		auther, err := d.store.Auth.Get(s.AuthMethod)
		checkErr(err)
		printSettings(s, auther)
	}, pythonConfig{}),
}

package cmd

import (
	"github.com/filebrowser/filebrowser/v2/storage"
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
	Run: python(func(cmd *cobra.Command, args []string, st *storage.Storage) {
		s, err := st.Settings.Get()
		checkErr(err)
		auther, err := st.Auth.Get(s.AuthMethod)
		checkErr(err)
		printSettings(s, auther)
	}, pythonConfig{}),
}

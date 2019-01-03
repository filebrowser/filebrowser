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
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()
		st := getFileBrowser(db)
		s := st.GetSettings()
		auther, err := st.GetAuther(s.AuthMethod)
		checkErr(err)
		printSettings(s, auther)
	},
}

package cmd

import (
	"github.com/filebrowser/filebrowser/storage/bolt/importer"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().String("old.db", "", "")
	importCmd.Flags().String("old.config", "", "")
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Imports an old configuration.",
	Long:  `Imports an old configuration`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		oldDB := mustGetString(cmd, "old.db")
		oldConf := mustGetString(cmd, "old.config")

		err := importer.Import(oldDB, oldConf, databasePath)
		checkErr(err)
	},
}

package cmd

import (
	"github.com/filebrowser/filebrowser/storage/bolt/importer"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().String("old.database", "", "")
	importCmd.Flags().String("old.config", "", "")
	importCmd.MarkFlagRequired("old.database")
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Imports an old configuration",
	Long:  `Imports an old configuration. This command DOES NOT
import share links because they are incompatible with
this version.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		oldDB := mustGetString(cmd, "old.database")
		oldConf := mustGetString(cmd, "old.config")

		err := importer.Import(oldDB, oldConf, databasePath)
		checkErr(err)
	},
}

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/filebrowser/filebrowser/v2/storage/bolt/importer"
)

func init() {
	rootCmd.AddCommand(upgradeCmd)

	upgradeCmd.Flags().String("old.database", "", "")
	upgradeCmd.Flags().String("old.config", "", "")
	_ = upgradeCmd.MarkFlagRequired("old.database")
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrades an old configuration",
	Long: `Upgrades an old configuration. This command DOES NOT
import share links because they are incompatible with
this version.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		flags := cmd.Flags()
		oldDB := mustGetString(flags, "old.database")
		oldConf := mustGetString(flags, "old.config")
		err := importer.Import(oldDB, oldConf, getParam(flags, "database"))
		checkErr(err)
	},
}

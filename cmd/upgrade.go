package cmd

import (
	"github.com/filebrowser/filebrowser/v2/storage/bolt/importer"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upgradeCmd)

	upgradeCmd.Flags().String("old.database", "", "")
	upgradeCmd.Flags().String("old.config", "", "")
	upgradeCmd.MarkFlagRequired("old.database")
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrades an old configuration",
	Long: `Upgrades an old configuration. This command DOES NOT
import share links because they are incompatible with
this version.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		oldDB := mustGetString(cmd, "old.database")
		oldConf := mustGetString(cmd, "old.config")

		err := importer.Import(oldDB, oldConf, mustGetStringViperFlag(cmd.Flags(), "database"))
		checkErr(err)
	},
}

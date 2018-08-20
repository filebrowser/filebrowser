package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dbCmd represents the db command
var dbCmd = &cobra.Command{
	Use:     "db",
	Version: rootCmd.Version,
	Aliases: []string{"database"},
	Short:   "Manage a filebrowser database",
	Long: `This is a CLI tool to ease the management of
filebrowser database files.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("db called. Command not implemented, yet.")
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
}

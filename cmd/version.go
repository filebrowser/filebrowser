package cmd

import (
	"fmt"

	"github.com/filebrowser/filebrowser/v2/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("File Browser v" + version.Version + "/" + version.CommitSHA)
	},
}

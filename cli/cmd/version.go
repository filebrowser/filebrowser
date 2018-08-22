package cmd

import (
	"text/template"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of File Browser",
	Long:  `All software has versions. This is File Browser's`,
	Run: func(cmd *cobra.Command, args []string) {
		// https://github.com/spf13/cobra/issues/724
		t := template.New("version")
		template.Must(t.Parse(rootCmd.VersionTemplate()))
		err := t.Execute(rootCmd.OutOrStdout(), rootCmd)
		if err != nil {
			rootCmd.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	serveCmd.AddCommand(versionCmd)
	dbCmd.AddCommand(versionCmd)
}

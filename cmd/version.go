package cmd

import (
	"text/template"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
	cmdsCmd.AddCommand(versionCmd)
	configCmd.AddCommand(versionCmd)
	hashCmd.AddCommand(versionCmd)
	upgradeCmd.AddCommand(versionCmd)
	rulesCmd.AddCommand(versionCmd)
	usersCmd.AddCommand(versionCmd)
}

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

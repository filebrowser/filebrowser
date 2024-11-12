package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rulesCmd.AddCommand(rulesLsCommand)
}

var rulesLsCommand = &cobra.Command{
	Use:   "ls",
	Short: "List global rules or user specific rules",
	Long:  `List global rules or user specific rules.`,
	Args:  cobra.NoArgs,
	Run: python(func(cmd *cobra.Command, _ []string, d pythonData) {
		runRules(d.store, cmd, nil, nil)
	}, pythonConfig{}),
}

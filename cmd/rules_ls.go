package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rulesCmd.AddCommand(rulesLsCommand)
}

var rulesLsCommand = &cobra.Command{
	Use:   "ls",
	Short: "List global rules or user specific rules",
	Long:  `List global rules or user specific rules.`,
	Args:  cobra.NoArgs,
	RunE: python(func(cmd *cobra.Command, _ []string, v *viper.Viper, d *pythonData) error {
		return runRules(d.store, cmd, v, nil, nil)
	}, pythonConfig{}),
}

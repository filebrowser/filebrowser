package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdsCmd)
}

var cmdsCmd = &cobra.Command{
	Use:   "cmds",
	Short: "Command runner management utility",
	Long:  `Command runner management utility.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

func printEvents(m map[string][]string) {
	for evt, cmds := range m {
		for i, cmd := range cmds {
			fmt.Printf("%s(%d): %s\n", evt, i, cmd)
		}
	}
}

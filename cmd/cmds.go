package cmd

import (
	"fmt"

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
}

func printEvents(m map[string][]string) {
	for evt, cmds := range m {
		for i, cmd := range cmds {
			fmt.Printf("%s(%d): %s\n", evt, i, cmd)
		}
	}
}

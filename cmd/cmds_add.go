package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	cmdsCmd.AddCommand(cmdsAddCmd)
}

var cmdsAddCmd = &cobra.Command{
	Use:   "add <event> <command>",
	Short: "Add a command to run on a specific event",
	Long:  `Add a command to run on a specific event.`,
	Args:  cobra.MinimumNArgs(2),
	RunE: python(func(_ *cobra.Command, args []string, d *pythonData) error {
		s, err := d.store.Settings.Get()
		if err != nil {
			return err
		}
		command := strings.Join(args[1:], " ")
		s.Commands[args[0]] = append(s.Commands[args[0]], command)
		err = d.store.Settings.Save(s)
		if err != nil {
			return err
		}
		printEvents(s.Commands)
		return nil
	}, pythonConfig{}),
}

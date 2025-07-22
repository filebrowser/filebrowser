package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	cmdsCmd.AddCommand(cmdsLsCmd)
	cmdsLsCmd.Flags().StringP("event", "e", "", "event name, without 'before' or 'after'")
}

var cmdsLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all commands for each event",
	Long:  `List all commands for each event.`,
	Args:  cobra.NoArgs,
	RunE: python(func(cmd *cobra.Command, _ []string, d *pythonData) error {
		s, err := d.store.Settings.Get()
		if err != nil {
			return err
		}
		evt, err := getString(cmd.Flags(), "event")
		if err != nil {
			return err
		}

		if evt == "" {
			printEvents(s.Commands)
		} else {
			show := map[string][]string{}
			show["before_"+evt] = s.Commands["before_"+evt]
			show["after_"+evt] = s.Commands["after_"+evt]
			printEvents(show)
		}
		return nil
	}, pythonConfig{}),
}

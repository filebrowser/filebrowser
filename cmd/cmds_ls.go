package cmd

import (
	"github.com/filebrowser/filebrowser/v2/storage"
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
	Run: python(func(cmd *cobra.Command, args []string, st *storage.Storage) {
		s, err := st.Settings.Get()
		checkErr(err)
		evt := mustGetString(cmd, "event")

		if evt == "" {
			printEvents(s.Commands)
		} else {
			show := map[string][]string{}
			show["before_"+evt] = s.Commands["before_"+evt]
			show["after_"+evt] = s.Commands["after_"+evt]
			printEvents(show)
		}
	}, pythonConfig{}),
}

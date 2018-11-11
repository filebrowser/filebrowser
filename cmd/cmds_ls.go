package cmd

import (
	"github.com/filebrowser/filebrowser/bolt"
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
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()
		st := bolt.ConfigStore{DB: db}
		r, err := st.GetRunner()
		checkErr(err)
		evt := mustGetString(cmd, "event")

		if evt == "" {
			printEvents(r.Commands)
		} else {
			show := map[string][]string{}
			show["before_"+evt] = r.Commands["before_"+evt]
			show["after_"+evt] = r.Commands["after_"+evt]
			printEvents(show)
		}
	},
}

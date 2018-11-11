package cmd

import (
	"github.com/filebrowser/filebrowser/bolt"
	"github.com/spf13/cobra"
)

func init() {
	cmdsCmd.AddCommand(cmdsAddCmd)
	cmdsAddCmd.Flags().StringP("command", "c", "", "command to add")
	cmdsAddCmd.Flags().StringP("event", "e", "", "corresponding event")
	cmdsAddCmd.MarkFlagRequired("command")
	cmdsAddCmd.MarkFlagRequired("event")
}

var cmdsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a command to run on a specific event",
	Long:  `Add a command to run on a specific event.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()
		st := bolt.ConfigStore{DB: db}
		r, err := st.GetRunner()
		checkErr(err)

		evt := mustGetString(cmd, "event")
		command := mustGetString(cmd, "command")

		r.Commands[evt] = append(r.Commands[evt], command)
		err = st.SaveRunner(r)
		checkErr(err)
		printEvents(r.Commands)
	},
}

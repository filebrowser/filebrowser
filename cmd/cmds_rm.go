package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	cmdsCmd.AddCommand(cmdsRmCmd)
	cmdsRmCmd.Flags().StringP("event", "e", "", "corresponding event")
	cmdsRmCmd.Flags().UintP("index", "i", 0, "command index")
	cmdsRmCmd.MarkFlagRequired("event")
	cmdsRmCmd.MarkFlagRequired("index")
}

var cmdsRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Removes a command from an event hooker",
	Long:  `Removes a command from an event hooker.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()
		st := getStore(db)
		s, err := st.Config.GetSettings()
		checkErr(err)

		evt := mustGetString(cmd, "event")
		i, err := cmd.Flags().GetUint("index")
		checkErr(err)

		s.Commands[evt] = append(s.Commands[evt][:i], s.Commands[evt][i+1:]...)
		err = st.Config.SaveSettings(s)
		checkErr(err)
		printEvents(s.Commands)
	},
}

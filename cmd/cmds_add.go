package cmd

import (
	"strings"

	"github.com/filebrowser/filebrowser/v2/storage"
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
	Run: python(func(cmd *cobra.Command, args []string, st *storage.Storage) {
		s, err := st.Settings.Get()
		checkErr(err)
		command := strings.Join(args[1:], " ")
		s.Commands[args[0]] = append(s.Commands[args[0]], command)
		err = st.Settings.Save(s)
		checkErr(err)
		printEvents(s.Commands)
	}, pythonConfig{}),
}

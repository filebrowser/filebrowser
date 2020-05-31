package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	cmdsCmd.AddCommand(cmdsRmCmd)
}

var cmdsRmCmd = &cobra.Command{
	Use:   "rm <event> <index> [index_end]",
	Short: "Removes a command from an event hooker",
	Long: `Removes a command from an event hooker. The provided index
is the same that's printed when you run 'cmds ls'. Note
that after each removal/addition, the index of the
commands change. So be careful when removing them after each
other.

You can also specify an optional parameter (index_end) so
you can remove all commands from 'index' to 'index_end',
including 'index_end'.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.RangeArgs(2, 3)(cmd, args); err != nil { //nolint:mnd
			return err
		}

		for _, arg := range args[1:] {
			if _, err := strconv.Atoi(arg); err != nil {
				return err
			}
		}

		return nil
	},
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		s, err := d.store.Settings.Get()
		checkErr(err)
		evt := args[0]

		i, err := strconv.Atoi(args[1])
		checkErr(err)
		f := i
		if len(args) == 3 { //nolint:mnd
			f, err = strconv.Atoi(args[2])
			checkErr(err)
		}

		s.Commands[evt] = append(s.Commands[evt][:i], s.Commands[evt][f+1:]...)
		err = d.store.Settings.Save(s)
		checkErr(err)
		printEvents(s.Commands)
	}, pythonConfig{}),
}

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
	Long:  `Removes a command from an event hooker.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.RangeArgs(2, 3)(cmd, args); err != nil {
			return err
		}

		for _, arg := range args[1:] {
			if _, err := strconv.Atoi(arg); err != nil {
				return err
			}
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()
		st := getStorage(db)
		s, err := st.Settings.Get()
		checkErr(err)
		evt := args[0]

		i, err := strconv.Atoi(args[1])
		checkErr(err)
		f := i
		if len(args) == 3 {
			f, err = strconv.Atoi(args[2])
			checkErr(err)
		}

		s.Commands[evt] = append(s.Commands[evt][:i], s.Commands[evt][f+1:]...)
		err = st.Settings.Save(s)
		checkErr(err)
		printEvents(s.Commands)
	},
}

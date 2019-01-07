package cmd

import (
	"strconv"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
)

func init() {
	rulesCmd.AddCommand(rulesRmCommand)
	rulesRmCommand.Flags().Uint("index", 0, "index of rule to remove")
	rulesRmCommand.MarkFlagRequired("index")
}

var rulesRmCommand = &cobra.Command{
	Use:   "rm <index> [index_end]",
	Short: "Remove a global rule or user rule",
	Long:  `Remove a global rule or user rule.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.RangeArgs(1, 2)(cmd, args); err != nil {
			return err
		}

		for _, arg := range args {
			if _, err := strconv.Atoi(arg); err != nil {
				return err
			}
		}

		return nil
	},
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		i, err := strconv.Atoi(args[0])
		checkErr(err)
		f := i
		if len(args) == 2 {
			f, err = strconv.Atoi(args[1])
			checkErr(err)
		}

		user := func(u *users.User) {
			u.Rules = append(u.Rules[:i], u.Rules[f+1:]...)
			err := d.store.Users.Save(u)
			checkErr(err)
		}

		global := func(s *settings.Settings) {
			s.Rules = append(s.Rules[:i], s.Rules[f+1:]...)
			err := d.store.Settings.Save(s)
			checkErr(err)
		}

		runRules(d.store, cmd, user, global)
	}, pythonConfig{}),
}

package cmd

import (
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
)

func init() {
	rulesCmd.AddCommand(rulesRmCommand)
	rulesRmCommand.Flags().Uint("index", 0, "index of rule to remove")
	rulesRmCommand.MarkFlagRequired("index")
}

var rulesRmCommand = &cobra.Command{
	Use:   "rm",
	Short: "Remove a global rule or user rule",
	Long:  `Remove a global rule or user rule.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		index := mustGetUint(cmd, "index")

		user := func(u *users.User, st *storage.Storage) {
			u.Rules = append(u.Rules[:index], u.Rules[index+1:]...)
			err := st.Users.Save(u)
			checkErr(err)
		}

		global := func(s *settings.Settings, st *storage.Storage) {
			s.Rules = append(s.Rules[:index], s.Rules[index+1:]...)
			err := st.Settings.Save(s)
			checkErr(err)
		}

		runRules(cmd, user, global)
	},
}

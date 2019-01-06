package cmd

import (
	"regexp"

	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
)

func init() {
	rulesCmd.AddCommand(rulesAddCmd)
	rulesAddCmd.Flags().BoolP("allow", "a", false, "indicates this is an allow rule")
	rulesAddCmd.Flags().BoolP("regex", "r", false, "indicates this is a regex rule")
}

var rulesAddCmd = &cobra.Command{
	Use:   "add <path|expression>",
	Short: "Add a global rule or user rule",
	Long:  `Add a global rule or user rule.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		allow := mustGetBool(cmd, "allow")
		regex := mustGetBool(cmd, "regex")
		exp := args[0]

		if regex {
			regexp.MustCompile(exp)
		}

		rule := rules.Rule{
			Allow: allow,
			Regex: regex,
		}

		if regex {
			rule.Regexp = &rules.Regexp{Raw: exp}
		} else {
			rule.Path = exp
		}

		user := func(u *users.User, st *storage.Storage) {
			u.Rules = append(u.Rules, rule)
			err := st.Users.Save(u)
			checkErr(err)
		}

		global := func(s *settings.Settings, st *storage.Storage) {
			s.Rules = append(s.Rules, rule)
			err := st.Settings.Save(s)
			checkErr(err)
		}

		runRules(cmd, user, global)
	},
}

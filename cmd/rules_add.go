package cmd

import (
	"regexp"

	"github.com/spf13/cobra"

	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
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
	RunE: python(func(cmd *cobra.Command, args []string, d *pythonData) error {
		allow, err := getBool(cmd.Flags(), "allow")
		if err != nil {
			return err
		}
		regex, err := getBool(cmd.Flags(), "regex")
		if err != nil {
			return err
		}
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

		user := func(u *users.User) error {
			u.Rules = append(u.Rules, rule)
			return d.store.Users.Save(u)
		}

		global := func(s *settings.Settings) error {
			s.Rules = append(s.Rules, rule)
			return d.store.Settings.Save(s)
		}

		return runRules(d.store, cmd, user, global)
	}, pythonConfig{}),
}

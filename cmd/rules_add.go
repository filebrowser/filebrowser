package cmd

import (
	"errors"
	"regexp"

	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
)

func init() {
	rulesCmd.AddCommand(rulesAddCmd)
	rulesAddCmd.Flags().BoolP("allow", "a", false, "allow rule instead of disallow")
	rulesAddCmd.Flags().StringP("path", "p", "", "path to which the rule applies")
	rulesAddCmd.Flags().StringP("regex", "r", "", "regex to which the rule applies")
}

var rulesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a global rule or user rule",
	Long: `Add a global rule or user rule. You must
set either path or regex.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		allow := mustGetBool(cmd, "allow")
		path := mustGetString(cmd, "path")
		regex := mustGetString(cmd, "regex")

		if path == "" && regex == "" {
			panic(errors.New("you must set either --path or --regex flags"))
		}

		if path != "" && regex != "" {
			panic(errors.New("you can't set --path and --regex flags at the same time"))
		}

		if regex != "" {
			regexp.MustCompile(regex)
		}

		rule := rules.Rule{
			Allow: allow,
			Path:  path,
			Regex: regex != "",
			Regexp: &rules.Regexp{
				Raw: regex,
			},
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

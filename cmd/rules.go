package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
)

func init() {
	rootCmd.AddCommand(rulesCmd)
	rulesCmd.PersistentFlags().StringP("username", "u", "", "username of user to which the rules apply")
	rulesCmd.PersistentFlags().UintP("id", "i", 0, "id of user to which the rules apply")
}

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Rules management utility",
	Long: `On each subcommand you'll have available at least two flags:
"username" and "id". You must either set only one of them
or none. If you set one of them, the command will apply to
an user, otherwise it will be applied to the global set or
rules.`,
	Args: cobra.NoArgs,
}

func runRules(st *storage.Storage, cmd *cobra.Command, usersFn func(*users.User), globalFn func(*settings.Settings)) {
	id := getUserIdentifier(cmd.Flags())
	if id != nil {
		user, err := st.Users.Get("", id)
		checkErr(err)

		if usersFn != nil {
			usersFn(user)
		}

		printRules(user.Rules, id)
		return
	}

	s, err := st.Settings.Get()
	checkErr(err)

	if globalFn != nil {
		globalFn(s)
	}

	printRules(s.Rules, id)
}

func getUserIdentifier(flags *pflag.FlagSet) interface{} {
	id := mustGetUint(flags, "id")
	username := mustGetString(flags, "username")

	if id != 0 {
		return id
	} else if username != "" {
		return username
	}

	return nil
}

func printRules(rulez []rules.Rule, id interface{}) {
	if id == nil {
		fmt.Printf("Global Rules:\n\n")
	} else {
		fmt.Printf("Rules for user %v:\n\n", id)
	}

	for id, rule := range rulez {
		fmt.Printf("(%d) ", id)
		if rule.Regex {
			if rule.Allow {
				fmt.Printf("Allow Regex: \t%s\n", rule.Regexp.Raw)
			} else {
				fmt.Printf("Disallow Regex: \t%s\n", rule.Regexp.Raw)
			}
		} else {
			if rule.Allow {
				fmt.Printf("Allow Path: \t%s\n", rule.Path)
			} else {
				fmt.Printf("Disallow Path: \t%s\n", rule.Path)
			}
		}
	}
}

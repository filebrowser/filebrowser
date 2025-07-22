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

func runRules(st *storage.Storage, cmd *cobra.Command, usersFn func(*users.User) error, globalFn func(*settings.Settings) error) error {
	id, err := getUserIdentifier(cmd.Flags())
	if err != nil {
		return err
	}
	if id != nil {
		var user *users.User
		user, err = st.Users.Get("", id)
		if err != nil {
			return err
		}

		if usersFn != nil {
			err = usersFn(user)
			if err != nil {
				return err
			}
		}

		printRules(user.Rules, id)
		return nil
	}

	s, err := st.Settings.Get()
	if err != nil {
		return err
	}

	if globalFn != nil {
		err = globalFn(s)
		if err != nil {
			return err
		}
	}

	printRules(s.Rules, id)
	return nil
}

func getUserIdentifier(flags *pflag.FlagSet) (interface{}, error) {
	id, err := getUint(flags, "id")
	if err != nil {
		return nil, err
	}
	username, err := getString(flags, "username")
	if err != nil {
		return nil, err
	}

	if id != 0 {
		return id, nil
	} else if username != "" {
		return username, nil
	}

	return nil, nil
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

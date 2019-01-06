package cmd

import (
	"strings"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	configCmd.AddCommand(configSetCmd)
	addConfigFlags(configSetCmd.Flags())
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Updates the configuration",
	Long: `Updates the configuration. Set the flags for the options
you want to change.`,
	Args: cobra.NoArgs,
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		s, err := d.store.Settings.Get()
		checkErr(err)

		hasAuth := false
		cmd.Flags().Visit(func(flag *pflag.Flag) {
			switch flag.Name {
			case "signup":
				s.Signup = mustGetBool(cmd, flag.Name)
			case "auth.method":
				hasAuth = true
			case "shell":
				s.Shell = strings.Split(strings.TrimSpace(mustGetString(cmd, flag.Name)), " ")
			case "branding.name":
				s.Branding.Name = mustGetString(cmd, flag.Name)
			case "branding.disableExternal":
				s.Branding.DisableExternal = mustGetBool(cmd, flag.Name)
			case "branding.files":
				s.Branding.Files = mustGetString(cmd, flag.Name)
			}
		})

		getUserDefaults(cmd, &s.Defaults, false)

		var auther auth.Auther
		var err error
		if hasAuth {
			s.AuthMethod, auther = getAuthentication(cmd)
			err = d.store.Auth.Save(auther)
			checkErr(err)
		} else {
			auther, err = d.store.Auth.Get(s.AuthMethod)
			checkErr(err)
		}

		err = d.store.Settings.Save(s)
		checkErr(err)
		printSettings(s, auther)
	}, pythonConfig{}),
}

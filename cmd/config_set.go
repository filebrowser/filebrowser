package cmd

import (
	"github.com/filebrowser/filebrowser/bolt"
	"github.com/filebrowser/filebrowser/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	configCmd.AddCommand(configSetCmd)
	addConfigFlags(configSetCmd)
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Updates the configuration",
	Long: `Updates the configuration. Set the flags for the options
you want to change.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()

		st := bolt.ConfigStore{DB: db}
		s, err := st.GetSettings()
		checkErr(err)

		auth := false
		cmd.Flags().Visit(func(flag *pflag.Flag) {
			switch flag.Name {
			case "baseURL":
				s.BaseURL = mustGetString(cmd, "baseURL")
			case "signup":
				s.Signup = mustGetBool(cmd, "signup")
			case "auth.method":
				auth = true
			case "branding.name":
				s.Branding.Name = mustGetString(cmd, "branding.name")
			case "branding.disableExternal":
				s.Branding.DisableExternal = mustGetBool(cmd, "branding.disableExternal")
			case "branding.files":
				s.Branding.Files = mustGetString(cmd, "branding.files")
			}
		})

		getUserDefaults(cmd, &s.Defaults, false)

		var auther types.Auther
		if auth {
			s.AuthMethod, auther = getAuthentication(cmd)
			err = st.SaveAuther(auther)
			checkErr(err)
		} else {
			auther, err = st.GetAuther(s.AuthMethod)
			checkErr(err)
		}

		err = st.SaveSettings(s)
		checkErr(err)
		printSettings(s, auther)
	},
}

package cmd

import (
	"strings"

	"github.com/filebrowser/filebrowser/v2/auth"
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

		st := getStorage(db)
		s, err := st.Settings.Get()
		checkErr(err)

		hasAuth := false
		cmd.Flags().Visit(func(flag *pflag.Flag) {
			switch flag.Name {
			case "baseURL":
				s.BaseURL = mustGetString(cmd, flag.Name)
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
			case "log":
				s.Log = mustGetString(cmd, flag.Name)
			case "address":
				s.Server.Address = mustGetString(cmd, flag.Name)
			case "port":
				s.Server.Port = mustGetInt(cmd, flag.Name)
			case "tls.cert":
				s.Server.TLSCert = mustGetString(cmd, flag.Name)
			case "tls.key":
				s.Server.TLSKey = mustGetString(cmd, flag.Name)
			}
		})

		getUserDefaults(cmd, &s.Defaults, false)

		var auther auth.Auther
		if hasAuth {
			s.AuthMethod, auther = getAuthentication(cmd)
			err = st.Auth.Save(auther)
			checkErr(err)
		} else {
			auther, err = st.Auth.Get(s.AuthMethod)
			checkErr(err)
		}

		err = st.Settings.Save(s)
		checkErr(err)
		printSettings(s, auther)
	},
}

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
		set, err := d.store.Settings.Get()
		checkErr(err)

		ser, err := d.store.Settings.GetServer()
		checkErr(err)

		hasAuth := false
		cmd.Flags().Visit(func(flag *pflag.Flag) {
			switch flag.Name {
			case "baseurl":
				ser.BaseURL = mustGetString(cmd, flag.Name)
			case "root":
				ser.Root = mustGetString(cmd, flag.Name)
			case "cert":
				ser.TLSCert = mustGetString(cmd, flag.Name)
			case "key":
				ser.TLSKey = mustGetString(cmd, flag.Name)
			case "address":
				ser.Address = mustGetString(cmd, flag.Name)
			case "port":
				ser.Port = mustGetString(cmd, flag.Name)
			case "log":
				ser.Log = mustGetString(cmd, flag.Name)
			case "signup":
				set.Signup = mustGetBool(cmd, flag.Name)
			case "auth.method":
				hasAuth = true
			case "shell":
				set.Shell = strings.Split(strings.TrimSpace(mustGetString(cmd, flag.Name)), " ")
			case "branding.name":
				set.Branding.Name = mustGetString(cmd, flag.Name)
			case "branding.disableExternal":
				set.Branding.DisableExternal = mustGetBool(cmd, flag.Name)
			case "branding.files":
				set.Branding.Files = mustGetString(cmd, flag.Name)
			}
		})

		getUserDefaults(cmd, &set.Defaults, false)

		var auther auth.Auther
		if hasAuth {
			set.AuthMethod, auther = getAuthentication(cmd)
			err = d.store.Auth.Save(auther)
			checkErr(err)
		} else {
			auther, err = d.store.Auth.Get(set.AuthMethod)
			checkErr(err)
		}

		err = d.store.Settings.Save(set)
		checkErr(err)
		err = d.store.Settings.SaveServer(ser)
		checkErr(err)
		printSettings(ser, set, auther)
	}, pythonConfig{}),
}

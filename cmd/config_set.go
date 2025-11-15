package cmd

import (
	"github.com/spf13/cobra"
	v "github.com/spf13/viper"
)

func init() {
	configCmd.AddCommand(configSetCmd)
	addConfigFlags(configSetCmd.Flags())
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Updates the configuration",
	Long: `Updates the configuration. Set the flags for the options
you want to change. Other options will remain unchanged.`,
	Args: cobra.NoArgs,
	RunE: python(func(cmd *cobra.Command, _ []string, d *pythonData) error {
		set, err := d.store.Settings.Get()
		if err != nil {
			return err
		}

		ser, err := d.store.Settings.GetServer()
		if err != nil {
			return err
		}

		hasAuth := false

		for _, key := range v.AllKeys() {
			if !v.IsSet(key) {
				continue
			}

			switch key {
			case "baseurl":
				ser.BaseURL = v.GetString(key)
			case "root":
				ser.Root = v.GetString(key)
			case "socket":
				ser.Socket = v.GetString(key)
			case "cert":
				ser.TLSCert = v.GetString(key)
			case "key":
				ser.TLSKey = v.GetString(key)
			case "address":
				ser.Address = v.GetString(key)
			case "port":
				ser.Port = v.GetString(key)
			case "log":
				ser.Log = v.GetString(key)
			case "hideloginbutton":
				set.HideLoginButton = v.GetBool(key)
			case "signup":
				set.Signup = v.GetBool(key)
			case "auth.method":
				hasAuth = true
			case "shell":
				var shell string
				shell = v.GetString(key)
				set.Shell = convertCmdStrToCmdArray(shell)
			case "createuserdir":
				set.CreateUserDir = v.GetBool(key)
			case "minimumpasswordlength":
				set.MinimumPasswordLength = v.GetUint(key)
			case "branding.name":
				set.Branding.Name = v.GetString(key)
			case "branding.color":
				set.Branding.Color = v.GetString(key)
			case "branding.theme":
				set.Branding.Theme = v.GetString(key)
			case "branding.disableexternal":
				set.Branding.DisableExternal = v.GetBool(key)
			case "branding.disableusedpercentage":
				set.Branding.DisableUsedPercentage = v.GetBool(key)
			case "branding.files":
				set.Branding.Files = v.GetString(key)
			case "filemode":
				set.FileMode, err = getAndParseMode(key)
			case "dirmode":
				set.DirMode, err = getAndParseMode(key)
			}

			if err != nil {
				return err
			}
		}

		err = getUserDefaults(&set.Defaults, false)
		if err != nil {
			return err
		}

		// read the defaults
		auther, err := d.store.Auth.Get(set.AuthMethod)
		if err != nil {
			return err
		}

		// check if there are new flags for existing auth method
		set.AuthMethod, auther, err = getAuthentication(hasAuth, set, auther)
		if err != nil {
			return err
		}

		err = d.store.Auth.Save(auther)
		if err != nil {
			return err
		}
		err = d.store.Settings.Save(set)
		if err != nil {
			return err
		}
		err = d.store.Settings.SaveServer(ser)
		if err != nil {
			return err
		}

		return printSettings(ser, set, auther)
	}, pythonConfig{}),
}

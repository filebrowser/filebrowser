package cmd

import (
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
you want to change. Other options will remain unchanged.`,
	Args: cobra.NoArgs,
	RunE: python(func(cmd *cobra.Command, _ []string, d *pythonData) error {
		flags := cmd.Flags()
		set, err := d.store.Settings.Get()
		if err != nil {
			return err
		}

		ser, err := d.store.Settings.GetServer()
		if err != nil {
			return err
		}

		hasAuth := false
		flags.Visit(func(flag *pflag.Flag) {
			if err != nil {
				return
			}
			switch flag.Name {
			case "baseurl":
				ser.BaseURL, err = mustGetString(flags, flag.Name)
			case "root":
				ser.Root, err = mustGetString(flags, flag.Name)
			case "socket":
				ser.Socket, err = mustGetString(flags, flag.Name)
			case "cert":
				ser.TLSCert, err = mustGetString(flags, flag.Name)
			case "key":
				ser.TLSKey, err = mustGetString(flags, flag.Name)
			case "address":
				ser.Address, err = mustGetString(flags, flag.Name)
			case "port":
				ser.Port, err = mustGetString(flags, flag.Name)
			case "log":
				ser.Log, err = mustGetString(flags, flag.Name)
			case "signup":
				set.Signup, err = mustGetBool(flags, flag.Name)
			case "auth.method":
				hasAuth = true
			case "shell":
				var shell string
				shell, err = mustGetString(flags, flag.Name)
				set.Shell = convertCmdStrToCmdArray(shell)
			case "create-user-dir":
				set.CreateUserDir, err = mustGetBool(flags, flag.Name)
			case "minimum-password-length":
				set.MinimumPasswordLength, err = mustGetUint(flags, flag.Name)
			case "branding.name":
				set.Branding.Name, err = mustGetString(flags, flag.Name)
			case "branding.color":
				set.Branding.Color, err = mustGetString(flags, flag.Name)
			case "branding.theme":
				set.Branding.Theme, err = mustGetString(flags, flag.Name)
			case "branding.disableExternal":
				set.Branding.DisableExternal, err = mustGetBool(flags, flag.Name)
			case "branding.disableUsedPercentage":
				set.Branding.DisableUsedPercentage, err = mustGetBool(flags, flag.Name)
			case "branding.files":
				set.Branding.Files, err = mustGetString(flags, flag.Name)
			case "file-mode":
				set.FileMode, err = mustGetMode(flags, flag.Name)
			case "dir-mode":
				set.DirMode, err = mustGetMode(flags, flag.Name)
			}
		})

		if err != nil {
			return err
		}

		err = getUserDefaults(flags, &set.Defaults, false)
		if err != nil {
			return err
		}

		// read the defaults
		auther, err := d.store.Auth.Get(set.AuthMethod)
		if err != nil {
			return err
		}

		// check if there are new flags for existing auth method
		set.AuthMethod, auther, err = getAuthentication(flags, hasAuth, set, auther)
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

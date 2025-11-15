package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	v "github.com/spf13/viper"

	"github.com/filebrowser/filebrowser/v2/settings"
)

func init() {
	configCmd.AddCommand(configInitCmd)
	addConfigFlags(configInitCmd.Flags())
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new database",
	Long: `Initialize a new database to use with File Browser. All of
this options can be changed in the future with the command
'filebrowser config set'. The user related flags apply
to the defaults when creating new users and you don't
override the options.`,
	Args: cobra.NoArgs,
	RunE: python(func(cmd *cobra.Command, _ []string, d *pythonData) error {
		defaults := settings.UserDefaults{}
		flags := cmd.Flags()
		err := getUserDefaults(&defaults, true)
		if err != nil {
			return err
		}
		authMethod, auther, err := getAuthentication(flags)
		if err != nil {
			return err
		}

		s := &settings.Settings{
			Key:                   generateKey(),
			Signup:                v.GetBool("signup"),
			HideLoginButton:       v.GetBool("hide-login-button"),
			CreateUserDir:         v.GetBool("create-user-dir"),
			MinimumPasswordLength: v.GetUint("minimum-password-length"),
			Shell:                 convertCmdStrToCmdArray(v.GetString("shell")),
			AuthMethod:            authMethod,
			Defaults:              defaults,
			Branding: settings.Branding{
				Name:                  v.GetString("branding.name"),
				DisableExternal:       v.GetBool("branding.disableexternal"),
				DisableUsedPercentage: v.GetBool("branding.disableusedpercentage"),
				Theme:                 v.GetString("branding.theme"),
				Files:                 v.GetString("branding.files"),
			},
		}

		s.FileMode, err = getAndParseMode("file-mode")
		if err != nil {
			return err
		}

		s.DirMode, err = getAndParseMode("dir-mode")
		if err != nil {
			return err
		}

		ser := &settings.Server{
			Address: v.GetString("address"),
			Socket:  v.GetString("socket"),
			Root:    v.GetString("root"),
			BaseURL: v.GetString("baseurl"),
			TLSKey:  v.GetString("key"),
			TLSCert: v.GetString("cert"),
			Port:    v.GetString("port"),
			Log:     v.GetString("log"),
		}

		err = d.store.Settings.Save(s)
		if err != nil {
			return err
		}
		err = d.store.Settings.SaveServer(ser)
		if err != nil {
			return err
		}
		err = d.store.Auth.Save(auther)
		if err != nil {
			return err
		}

		fmt.Printf(`
Congratulations! You've set up your database to use with File Browser.
Now add your first user via 'filebrowser users add' and then you just
need to call the main command to boot up the server.
`)
		return printSettings(ser, s, auther)
	}, pythonConfig{noDB: true}),
}

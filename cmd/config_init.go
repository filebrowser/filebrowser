package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

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
		err := getUserDefaults(flags, &defaults, true)
		if err != nil {
			return err
		}
		authMethod, auther, err := getAuthentication(flags)
		if err != nil {
			return err
		}

		key := generateKey()

		signup, err := mustGetBool(flags, "signup")
		if err != nil {
			return err
		}

		createUserDir, err := mustGetBool(flags, "create-user-dir")
		if err != nil {
			return err
		}

		minLength, err := mustGetUint(flags, "minimum-password-length")
		if err != nil {
			return err
		}

		shell, err := mustGetString(flags, "shell")
		if err != nil {
			return err
		}

		brandingName, err := mustGetString(flags, "branding.name")
		if err != nil {
			return err
		}

		brandingDisableExternal, err := mustGetBool(flags, "branding.disableExternal")
		if err != nil {
			return err
		}

		brandingDisableUsedPercentage, err := mustGetBool(flags, "branding.disableUsedPercentage")
		if err != nil {
			return err
		}

		brandingTheme, err := mustGetString(flags, "branding.theme")
		if err != nil {
			return err
		}

		brandingFiles, err := mustGetString(flags, "branding.files")
		if err != nil {
			return err
		}

		s := &settings.Settings{
			Key:                   key,
			Signup:                signup,
			CreateUserDir:         createUserDir,
			MinimumPasswordLength: minLength,
			Shell:                 convertCmdStrToCmdArray(shell),
			AuthMethod:            authMethod,
			Defaults:              defaults,
			Branding: settings.Branding{
				Name:                  brandingName,
				DisableExternal:       brandingDisableExternal,
				DisableUsedPercentage: brandingDisableUsedPercentage,
				Theme:                 brandingTheme,
				Files:                 brandingFiles,
			},
		}

		address, err := mustGetString(flags, "address")
		if err != nil {
			return err
		}

		socket, err := mustGetString(flags, "socket")
		if err != nil {
			return err
		}

		root, err := mustGetString(flags, "root")
		if err != nil {
			return err
		}

		baseURL, err := mustGetString(flags, "baseurl")
		if err != nil {
			return err
		}

		tlsKey, err := mustGetString(flags, "key")
		if err != nil {
			return err
		}

		cert, err := mustGetString(flags, "cert")
		if err != nil {
			return err
		}

		port, err := mustGetString(flags, "port")
		if err != nil {
			return err
		}

		log, err := mustGetString(flags, "log")
		if err != nil {
			return err
		}

		ser := &settings.Server{
			Address: address,
			Socket:  socket,
			Root:    root,
			BaseURL: baseURL,
			TLSKey:  tlsKey,
			TLSCert: cert,
			Port:    port,
			Log:     log,
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

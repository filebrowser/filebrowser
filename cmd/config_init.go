package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/filebrowser/filebrowser/v2/auth"
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
		flags := cmd.Flags()

		// General Settings
		s := &settings.Settings{
			Key: generateKey(),
		}

		err := getUserDefaults(flags, &s.Defaults, true)
		if err != nil {
			return err
		}

		s.Signup, err = flags.GetBool("signup")
		if err != nil {
			return err
		}

		s.HideLoginButton, err = flags.GetBool("hideLoginButton")
		if err != nil {
			return err
		}

		s.CreateUserDir, err = flags.GetBool("createUserDir")
		if err != nil {
			return err
		}

		s.MinimumPasswordLength, err = flags.GetUint("minimumPasswordLength")
		if err != nil {
			return err
		}

		shell, err := flags.GetString("shell")
		if err != nil {
			return err
		}
		s.Shell = convertCmdStrToCmdArray(shell)

		s.FileMode, err = getAndParseFileMode(flags, "fileMode")
		if err != nil {
			return err
		}

		s.DirMode, err = getAndParseFileMode(flags, "dirMode")
		if err != nil {
			return err
		}

		s.Branding.Name, err = flags.GetString("branding.name")
		if err != nil {
			return err
		}

		s.Branding.DisableExternal, err = flags.GetBool("branding.disableExternal")
		if err != nil {
			return err
		}

		s.Branding.DisableUsedPercentage, err = flags.GetBool("branding.disableUsedPercentage")
		if err != nil {
			return err
		}

		s.Branding.Theme, err = flags.GetString("branding.themes")
		if err != nil {
			return err
		}

		s.Branding.Files, err = flags.GetString("branding.files")
		if err != nil {
			return err
		}

		s.Tus.ChunkSize, err = flags.GetUint64("tus.chunkSize")
		if err != nil {
			return err
		}

		s.Tus.RetryCount, err = flags.GetUint16("tus.retryCount")
		if err != nil {
			return err
		}

		var auther auth.Auther
		s.AuthMethod, auther, err = getAuthentication(flags)
		if err != nil {
			return err
		}

		// Server Settings
		ser := &settings.Server{}
		ser.Address, err = flags.GetString("address")
		if err != nil {
			return err
		}

		ser.Socket, err = flags.GetString("socket")
		if err != nil {
			return err
		}

		ser.Root, err = flags.GetString("root")
		if err != nil {
			return err
		}

		ser.BaseURL, err = flags.GetString("baseURL")
		if err != nil {
			return err
		}

		ser.TLSKey, err = flags.GetString("key")
		if err != nil {
			return err
		}

		ser.TLSCert, err = flags.GetString("cert")
		if err != nil {
			return err
		}

		ser.Port, err = flags.GetString("port")
		if err != nil {
			return err
		}

		ser.Log, err = flags.GetString("log")
		if err != nil {
			return err
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
	}, pythonConfig{expectsNoDatabase: true}),
}

package cmd

import (
	filebrowser "github.com/filebrowser/filebrowser/lib"
	"github.com/spf13/cobra"
	v "github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:     "serve",
	Version: rootCmd.Version,
	Aliases: []string{"server"},
	Short:   "Start filebrowser service",
	Long:    rootCmd.Long,
	Run: func(cmd *cobra.Command, args []string) {
		Serve()
	},
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	f := serveCmd.PersistentFlags()

	flag := func(k string, i interface{}, u string) {
		switch y := i.(type) {
		case bool:
			f.Bool(k, y, u)
		case int:
			f.Int(k, y, u)
		case string:
			f.String(k, y, u)
		}
		v.SetDefault(k, i)
	}

	flagP := func(k, p string, i interface{}, u string) {
		switch y := i.(type) {
		case bool:
			f.BoolP(k, p, y, u)
		case int:
			f.IntP(k, p, y, u)
		case string:
			f.StringP(k, p, y, u)
		}
		v.SetDefault(k, i)
	}

	deprecated := func(k string, i interface{}, u, m string) {
		switch y := i.(type) {
		case bool:
			f.Bool(k, y, u)
		case int:
			f.Int(k, y, u)
		case string:
			f.String(k, y, u)
		}
		f.MarkDeprecated(k, m)
	}

	// Global settings
	flagP("port", "p", 0, "HTTP Port (default is random)")
	flagP("address", "a", "", "Address to listen to (default is all of them)")
	flagP("database", "d", "./filebrowser.db", "Database file")
	flagP("log", "l", "stdout", "Errors logger; can use 'stdout', 'stderr' or file")
	flagP("baseurl", "b", "", "Base URL")
	flag("prefixurl", "", "Prefix URL")
	flag("staticgen", "", "Static Generator you want to enable")

	// User default settings
	f.String("defaults.commands", "git svn hg", "Default commands option for new users")
	v.SetDefault("defaults.commands", []string{"git", "svn", "hg"})

	flagP("defaults.scope", "s", ".", "Default scope option for new users")
	flag("defaults.viewMode", filebrowser.MosaicViewMode, "Default view mode for new users")
	flag("defaults.allowCommands", true, "Default allow commands option for new users")
	flag("defaults.allowEdit", true, "Default allow edit option for new users")
	flag("defaults.allowNew", true, "Default allow new option for new users")
	flag("defaults.allowPublish", true, "Default allow publish option for new users")
	flag("defaults.locale", "", "Default locale for new users, set it empty to enable auto detect from browser")

	// Recaptcha settings
	flag("recaptcha.host", "https://www.google.com", "Use another host for ReCAPTCHA. recaptcha.net might be useful in China")
	flag("recaptcha.key", "", "ReCaptcha site key")
	flag("recaptcha.secret", "", "ReCaptcha secret")

	// Auth settings
	flag("auth.method", "default", "Switch between 'none', 'default' and 'proxy' authentication")
	flag("auth.header", "X-Forwarded-User", "The header name used for proxy authentication")

	// Bind the full flag set to the configuration
	if err := v.BindPFlags(f); err != nil {
		panic(err)
	}

	// Deprecated flags
	deprecated("no-auth", false, "Disables authentication", "use --auth.method='none' instead")
	deprecated("alternative-recaptcha", false, "Use recaptcha.net for serving and handling, useful in China", "use --recaptcha.host instead")
	deprecated("recaptcha-key", "", "ReCaptcha site key", "use --recaptcha.key instead")
	deprecated("recaptcha-secret", "", "ReCaptcha secret", "use --recaptcha.secret instead")
	deprecated("scope", ".", "Default scope option for new users", "use --defaults.scope instead")
	deprecated("commands", "git svn hg", "Default commands option for new users", "use --defaults.commands instead")
	deprecated("view-mode", "mosaic", "Default view mode for new users", "use --defaults.viewMode instead")
	deprecated("locale", "", "Default locale for new users, set it empty to enable auto detect from browser", "use --defaults.locale instead")
	deprecated("allow-commands", true, "Default allow commands option for new users", "use --defaults.allowCommands instead")
	deprecated("allow-edit", true, "Default allow edit option for new users", "use --defaults.allowEdit instead")
	deprecated("allow-publish", true, "Default allow publish option for new users", "use --defaults.allowPublish instead")
	deprecated("allow-new", true, "Default allow new option for new users", "use --defaults.allowNew instead")
}

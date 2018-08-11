package cmd

import (
	filebrowser "github.com/filebrowser/filebrowser/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v "github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:     "serve",
	Version: rootCmd.Version,
	Aliases: []string{"server"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		Serve()
	},
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	f := serveCmd.PersistentFlags()

	// Global settings
	port := 0
	addr := ""
	database := "./filebrowser.db"
	logfile := "stdout"
	baseurl := ""
	prefixurl := ""
	staticg := ""

	f.IntP("port", "p", port, "HTTP Port (default is random)")
	f.StringP("address", "a", addr, "Address to listen to (default is all of them)")
	f.StringP("database", "d", database, "Database file")
	f.StringP("log", "l", logfile, "Errors logger; can use 'stdout', 'stderr' or file")
	f.StringP("baseurl", "b", baseurl, "Base URL")
	f.String("prefixurl", prefixurl, "Prefix URL")
	f.String("staticgen", staticg, "Static Generator you want to enable")

	v.SetDefault("port", port)
	v.SetDefault("address", addr)
	v.SetDefault("database", database)
	v.SetDefault("log", logfile)
	v.SetDefault("baseurl", baseurl)
	v.SetDefault("prefixurl", prefixurl)
	v.SetDefault("staticgen", staticg)

	// User default settings
	var defaults = struct {
		commands      string
		scope         string
		viewMode      string
		allowCommands bool
		allowEdit     bool
		allowNew      bool
		allowPublish  bool
		locale        string
	}{
		"git svn hg",
		".",
		filebrowser.MosaicViewMode,
		true,
		true,
		true,
		true,
		"",
	}

	f.String("defaults.commands", defaults.commands, "Default commands option for new users")
	f.StringP("defaults.scope", "s", defaults.scope, "Default scope option for new users")
	f.String("defaults.viewMode", defaults.viewMode, "Default view mode for new users")
	f.Bool("defaults.allowCommands", defaults.allowCommands, "Default allow commands option for new users")
	f.Bool("defaults.allowEdit", defaults.allowEdit, "Default allow edit option for new users")
	f.Bool("defaults.allowNew", defaults.allowNew, "Default allow new option for new users")
	f.Bool("defaults.allowPublish", defaults.allowPublish, "Default allow publish option for new users")
	f.String("defaults.locale", defaults.locale, "Default locale for new users, set it empty to enable auto detect from browser")

	v.SetDefault("defaults.scope", defaults.scope)
	v.SetDefault("defaults.commands", []string{"git", "svn", "hg"})
	v.SetDefault("defaults.viewMode", defaults.viewMode)
	v.SetDefault("defaults.allowCommands", defaults.allowCommands)
	v.SetDefault("defaults.allowEdit", defaults.allowEdit)
	v.SetDefault("defaults.allowNew", defaults.allowNew)
	v.SetDefault("defaults.allowPublish", defaults.allowPublish)
	v.SetDefault("defaults.locale", defaults.locale)

	// Recaptcha settings
	var recaptcha = struct {
		host   string
		key    string
		secret string
	}{
		"https://www.google.com",
		"",
		"",
	}

	f.String("recaptcha.host", recaptcha.host, "Use another host for ReCAPTCHA. recaptcha.net might be useful in China")
	f.String("recaptcha.key", recaptcha.key, "ReCaptcha site key")
	f.String("recaptcha.secret", recaptcha.secret, "ReCaptcha secret")

	v.SetDefault("recaptcha.host", recaptcha.host)
	v.SetDefault("recaptcha.key", recaptcha.key)
	v.SetDefault("recaptcha.secret", recaptcha.secret)

	// Auth settings
	var auth = struct {
		method string
		header string
	}{
		"default",
		"X-Forwarded-User",
	}

	f.String("auth.method", auth.method, "Switch between 'none', 'default' and 'proxy' authentication")
	f.String("auth.header", auth.header, "The header name used for proxy authentication")

	v.SetDefault("auth.method", auth.method)
	v.SetDefault("auth.header", auth.header)

	// Bind the full flag set to the configuration
	if err := v.BindPFlags(f); err != nil {
		panic(err)
	}

	// Deprecated flags
	Deprecated(f, "no-auth", false, "Disables authentication", "use --auth.method='none' instead")
	Deprecated(f, "alternative-recaptcha", false, "Use recaptcha.net for serving and handling, useful in China", "use --recaptcha.host instead")
	Deprecated(f, "recaptcha-key", "", "ReCaptcha site key", "use --recaptcha.key instead")
	Deprecated(f, "recaptcha-secret", "", "ReCaptcha secret", "use --recaptcha.secret instead")
	Deprecated(f, "scope", ".", "Default scope option for new users", "use --defaults.scope instead")
	Deprecated(f, "commands", "git svn hg", "Default commands option for new users", "use --defaults.commands instead")
	Deprecated(f, "view-mode", "mosaic", "Default view mode for new users", "use --defaults.viewMode instead")
	Deprecated(f, "locale", "", "Default locale for new users, set it empty to enable auto detect from browser", "use --defaults.locale instead")
	Deprecated(f, "allow-commands", true, "Default allow commands option for new users", "use --defaults.allowCommands instead")
	Deprecated(f, "allow-edit", true, "Default allow edit option for new users", "use --defaults.allowEdit instead")
	Deprecated(f, "allow-publish", true, "Default allow publish option for new users", "use --defaults.allowPublish instead")
	Deprecated(f, "allow-new", true, "Default allow new option for new users", "use --defaults.allowNew instead")
}

func Deprecated(f *pflag.FlagSet, k string, i interface{}, u, m string) {
	switch v := i.(type) {
	case bool:
		f.Bool(k, v, u)
	case string:
		f.String(k, v, u)
	}
	f.MarkDeprecated(k, m)
}

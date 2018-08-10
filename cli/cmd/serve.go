package cmd

import (
	"github.com/spf13/cobra"
	v "github.com/spf13/viper"
	filebrowser "github.com/filebrowser/filebrowser/lib"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Version: rootCmd.Version,
	Aliases: []string{"server"},
	Short: "A brief description of your command",
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
	l := f.Lookup

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

	v.SetDefault("Port", port)
	v.SetDefault("Address", addr)
	v.SetDefault("Database", database)
	v.SetDefault("Logger", logfile)
	v.SetDefault("BaseURL", baseurl)
	v.SetDefault("PrefixURL", prefixurl)
	v.SetDefault("StaticGen", staticg)

	v.BindPFlag("Port", l("port"))
	v.BindPFlag("Address", l("address"))
	v.BindPFlag("Database", l("database"))
	v.BindPFlag("Logger", l("log"))
	v.BindPFlag("BaseURL", l("baseurl"))
	v.BindPFlag("PrefixURL", l("prefixurl"))
	v.BindPFlag("StaticGen", l("staticgen"))

	// User default settings
	var defaults = struct {
		commands string
		scope string
		viewMode string
		allowCommands bool
		allowEdit bool
		allowNew bool
		allowPublish bool
		locale string
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

	v.SetDefault("Defaults.Scope", defaults.scope)
	v.SetDefault("Defaults.Commands", []string{"git", "svn", "hg"})
	v.SetDefault("Defaults.ViewMode", defaults.viewMode)
	v.SetDefault("Defaults.AllowCommmands", defaults.allowCommands)
	v.SetDefault("Defaults.AllowEdit", defaults.allowEdit)
	v.SetDefault("Defaults.AllowNew", defaults.allowNew)
	v.SetDefault("Defaults.AllowPublish", defaults.allowPublish)
	v.SetDefault("Defaults.Locale", defaults.locale)

	v.BindPFlag("Defaults.Scope", l("defaults.scope"))
	v.BindPFlag("Defaults.Commands", l("defaults.commands"))
	v.BindPFlag("Defaults.ViewMode", l("defaults.viewMode"))
	v.BindPFlag("Defaults.AllowCommands", l("defaults.allowCommands"))
	v.BindPFlag("Defaults.AllowEdit", l("defaults.allowEdit"))
	v.BindPFlag("Defaults.AllowNew", l("defaults.allowNew"))
	v.BindPFlag("Defaults.AllowPublish", l("defaults.allowPublish"))
	v.BindPFlag("Defaults.Locale", l("defaults.locale"))

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

	v.SetDefault("Recaptcha.Host", recaptcha.host)
	v.SetDefault("Recaptcha.Key", recaptcha.key)
	v.SetDefault("Recaptcha.Secret", recaptcha.secret)

	v.BindPFlag("Recaptcha.Host", l("recaptcha.host"))
	v.BindPFlag("Recaptcha.Key", l("recaptcha.key"))
	v.BindPFlag("Recaptcha.Secret", l("recaptcha.secret"))

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

	v.SetDefault("Auth.Method", auth.method)
	v.SetDefault("Auth.Header", auth.header)

	v.BindPFlag("Auth.Method", l("auth.method"))
	v.BindPFlag("Auth.Header", l("auth.header"))
}

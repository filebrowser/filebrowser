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

var (
	addr      string
	database  string
	scope     string
	commands  string
	logfile   string
	staticg   string
	locale    string
	baseurl   string
	prefixurl string
	viewMode  string
	port      int
	recaptcha struct {
		host   string
		key    string
		secret string
	}
	auth struct {
		method string
		header string
	}
	allowCommands bool
	allowEdit     bool
	allowNew      bool
	allowPublish  bool
	showVer       bool
)

func init() {
	rootCmd.AddCommand(serveCmd)

	f := serveCmd.PersistentFlags()
	l := f.Lookup

	f.IntVarP(&port, "port", "p", 0, "HTTP Port (default is random)")
	f.StringVarP(&addr, "address", "a", "", "Address to listen to (default is all of them)")
	f.StringVarP(&database, "database", "d", "./filebrowser.db", "Database file")
	f.StringVarP(&logfile, "log", "l", "stdout", "Errors logger; can use 'stdout', 'stderr' or file")
	f.StringVarP(&baseurl, "baseurl", "b", "", "Base URL")
	f.StringVar(&prefixurl, "prefixurl", "", "Prefix URL")
	f.StringVar(&staticg, "staticgen", "", "Static Generator you want to enable")

	// User default values
	f.StringVar(&commands, "defaults.commands", "git svn hg", "Default commands option for new users")
	f.StringVarP(&scope, "defaults.scope", "s", ".", "Default scope option for new users")
	f.StringVar(&viewMode, "defaults.viewMode", "mosaic", "Default view mode for new users")
	f.BoolVar(&allowCommands, "defaults.allowCommands", true, "Default allow commands option for new users")
	f.BoolVar(&allowEdit, "defaults.allowEdit", true, "Default allow edit option for new users")
	f.BoolVar(&allowPublish, "defaults.allowPublish", true, "Default allow publish option for new users")
	f.BoolVar(&allowNew, "defaults.allowNew", true, "Default allow new option for new users")
	f.StringVar(&locale, "defaults.locale", "", "Default locale for new users, set it empty to enable auto detect from browser")

	// Recaptcha settings
	f.StringVar(&recaptcha.host, "recaptcha.host", "https://www.google.com", "Use another host for ReCAPTCHA. recaptcha.net might be useful in China")
	f.StringVar(&recaptcha.key, "recaptcha.key", "", "ReCaptcha site key")
	f.StringVar(&recaptcha.secret, "recaptcha.secret", "", "ReCaptcha secret")

	// Auth settings
	f.StringVar(&auth.method, "auth.method", "default", "Switch between 'none', 'default' and 'proxy' authentication")
	f.StringVar(&auth.header, "auth.header", "X-Forwarded-User", "The header name used for proxy authentication")

	v.SetDefault("Port", "0")
	v.SetDefault("Address", "")
	v.SetDefault("Database", "./filebrowser.db")
	v.SetDefault("Logger", "stdout")
	v.SetDefault("BaseURL", "")
	v.SetDefault("PrefixURL", "")
	v.SetDefault("StaticGen", "")

	v.BindPFlag("Port", l("port"))
	v.BindPFlag("Address", l("address"))
	v.BindPFlag("Database", l("database"))
	v.BindPFlag("Logger", l("log"))
	v.BindPFlag("BaseURL", l("baseurl"))
	v.BindPFlag("PrefixURL", l("prefixurl"))
	v.BindPFlag("StaticGen", l("staticgen"))

	// User default values
	v.SetDefault("Defaults.Scope", ".")
	v.SetDefault("Defaults.Commands", []string{"git", "svn", "hg"})
	v.SetDefault("Defaults.ViewMode", filebrowser.MosaicViewMode)
	v.SetDefault("Defaults.AllowCommmands", true)
	v.SetDefault("Defaults.AllowEdit", true)
	v.SetDefault("Defaults.AllowNew", true)
	v.SetDefault("Defaults.AllowPublish", true)
	v.SetDefault("Defaults.Locale", "")

	v.BindPFlag("Defaults.Scope", l("defaults.scope"))
	v.BindPFlag("Defaults.Commands", l("defaults.commands"))
	v.BindPFlag("Defaults.ViewMode", l("defaults.viewMode"))
	v.BindPFlag("Defaults.AllowCommands", l("defaults.allowCommands"))
	v.BindPFlag("Defaults.AllowEdit", l("defaults.allowEdit"))
	v.BindPFlag("Defaults.AllowNew", l("defaults.allowNew"))
	v.BindPFlag("Defaults.AllowPublish", l("defaults.allowPublish"))
	v.BindPFlag("Defaults.Locale", l("defaults.locale"))

	// Recaptcha settings
	v.SetDefault("Recaptcha.Host", "https://www.google.com")
	v.SetDefault("Recaptcha.Key", "")
	v.SetDefault("Recaptcha.Secret", "")

	v.BindPFlag("Recaptcha.Host", l("recaptcha.host"))
	v.BindPFlag("Recaptcha.Key", l("recaptcha.key"))
	v.BindPFlag("Recaptcha.Secret", l("recaptcha.secret"))

	// Auth settings
	v.SetDefault("Auth.Method", "default")
	v.SetDefault("Auth.Header", "X-Fowarded-User")

	v.BindPFlag("Auth.Method", l("auth.method"))
	v.BindPFlag("Auth.Header", l("auth.header"))
}

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	serveCmd.PersistentFlags().IntVarP(&port, "port", "p", 0, "HTTP Port (default is random)")
	serveCmd.PersistentFlags().StringVarP(&addr, "address", "a", "", "Address to listen to (default is all of them)")
	serveCmd.PersistentFlags().StringVarP(&database, "database", "d", "./filebrowser.db", "Database file")
	serveCmd.PersistentFlags().StringVarP(&logfile, "log", "l", "stdout", "Errors logger; can use 'stdout', 'stderr' or file")
	serveCmd.PersistentFlags().StringVarP(&baseurl, "baseurl", "b", "", "Base URL")
	serveCmd.PersistentFlags().StringVar(&prefixurl, "prefixurl", "", "Prefix URL")
	serveCmd.PersistentFlags().StringVar(&staticg, "staticgen", "", "Static Generator you want to enable")

	// User default values
	serveCmd.PersistentFlags().StringVar(&commands, "defaults.commands", "git svn hg", "Default commands option for new users")
	serveCmd.PersistentFlags().StringVarP(&scope, "defaults.scope", "s", ".", "Default scope option for new users")
	serveCmd.PersistentFlags().StringVar(&viewMode, "defaults.viewMode", "mosaic", "Default view mode for new users")
	serveCmd.PersistentFlags().BoolVar(&allowCommands, "defaults.allowCommands", true, "Default allow commands option for new users")
	serveCmd.PersistentFlags().BoolVar(&allowEdit, "defaults.allowEdit", true, "Default allow edit option for new users")
	serveCmd.PersistentFlags().BoolVar(&allowPublish, "defaults.allowPublish", true, "Default allow publish option for new users")
	serveCmd.PersistentFlags().BoolVar(&allowNew, "defaults.allowNew", true, "Default allow new option for new users")
	serveCmd.PersistentFlags().StringVar(&locale, "defaults.locale", "", "Default locale for new users, set it empty to enable auto detect from browser")

	// Recaptcha settings
	serveCmd.PersistentFlags().StringVar(&recaptcha.host, "recaptcha.host", "https://www.google.com", "Use another host for ReCAPTCHA. recaptcha.net might be useful in China")
	serveCmd.PersistentFlags().StringVar(&recaptcha.key, "recaptcha.key", "", "ReCaptcha site key")
	serveCmd.PersistentFlags().StringVar(&recaptcha.secret, "recaptcha.secret", "", "ReCaptcha secret")

	// Auth settings
	serveCmd.PersistentFlags().StringVar(&auth.method, "auth.method", "default", "Switch between 'none', 'default' and 'proxy' authentication")
	serveCmd.PersistentFlags().StringVar(&auth.header, "auth.header", "X-Forwarded-User", "The header name used for proxy authentication")

	viper.SetDefault("Port", "0")
	viper.SetDefault("Address", "")
	viper.SetDefault("Database", "./filebrowser.db")
	viper.SetDefault("Logger", "stdout")
	viper.SetDefault("BaseURL", "")
	viper.SetDefault("PrefixURL", "")
	viper.SetDefault("StaticGen", "")

	viper.BindPFlag("Port", serveCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("Address", serveCmd.PersistentFlags().Lookup("address"))
	viper.BindPFlag("Database", serveCmd.PersistentFlags().Lookup("database"))
	viper.BindPFlag("Logger", serveCmd.PersistentFlags().Lookup("log"))
	viper.BindPFlag("BaseURL", serveCmd.PersistentFlags().Lookup("baseurl"))
	viper.BindPFlag("PrefixURL", serveCmd.PersistentFlags().Lookup("prefixurl"))
	viper.BindPFlag("StaticGen", serveCmd.PersistentFlags().Lookup("staticgen"))

	// User default values
	viper.SetDefault("Defaults.Scope", ".")
	viper.SetDefault("Defaults.Commands", []string{"git", "svn", "hg"})
	viper.SetDefault("Defaults.ViewMode", filebrowser.MosaicViewMode)
	viper.SetDefault("Defaults.AllowCommmands", true)
	viper.SetDefault("Defaults.AllowEdit", true)
	viper.SetDefault("Defaults.AllowNew", true)
	viper.SetDefault("Defaults.AllowPublish", true)
	viper.SetDefault("Defaults.Locale", "")

	viper.BindPFlag("Defaults.Scope", serveCmd.PersistentFlags().Lookup("defaults.scope"))
	viper.BindPFlag("Defaults.Commands", serveCmd.PersistentFlags().Lookup("defaults.commands"))
	viper.BindPFlag("Defaults.ViewMode", serveCmd.PersistentFlags().Lookup("defaults.viewMode"))
	viper.BindPFlag("Defaults.AllowCommands", serveCmd.PersistentFlags().Lookup("defaults.allowCommands"))
	viper.BindPFlag("Defaults.AllowEdit", serveCmd.PersistentFlags().Lookup("defaults.allowEdit"))
	viper.BindPFlag("Defaults.AllowNew", serveCmd.PersistentFlags().Lookup("defaults.allowNew"))
	viper.BindPFlag("Defaults.AllowPublish", serveCmd.PersistentFlags().Lookup("defaults.allowPublish"))
	viper.BindPFlag("Defaults.Locale", serveCmd.PersistentFlags().Lookup("defaults.locale"))

	// Recaptcha settings
	viper.SetDefault("Recaptcha.Host", "https://www.google.com")
	viper.SetDefault("Recaptcha.Key", "")
	viper.SetDefault("Recaptcha.Secret", "")

	viper.BindPFlag("Recaptcha.Host", serveCmd.PersistentFlags().Lookup("recaptcha.host"))
	viper.BindPFlag("Recaptcha.Key", serveCmd.PersistentFlags().Lookup("recaptcha.key"))
	viper.BindPFlag("Recaptcha.Secret", serveCmd.PersistentFlags().Lookup("recaptcha.secret"))

	// Auth settings
	viper.SetDefault("Auth.Method", "default")
	viper.SetDefault("Auth.Header", "X-Fowarded-User")

	viper.BindPFlag("Auth.Method", serveCmd.PersistentFlags().Lookup("auth.method"))
	viper.BindPFlag("Auth.Header", serveCmd.PersistentFlags().Lookup("auth.header"))
}

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	filebrowser "github.com/filebrowser/filebrowser/lib"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Aliases: []string{"serve","server"},
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		Serve()
	},
}

var (
	addr            string
	database        string
	scope           string
	commands        string
	logfile         string
	staticg         string
	locale          string
	baseurl         string
	prefixurl       string
	viewMode        string
	recaptchakey    string
	recaptchasecret string
	port            int
	auth            struct {
		method      string
		loginHeader string
	}
	noAuth         bool
	allowCommands  bool
	allowEdit      bool
	allowNew       bool
	allowPublish   bool
	showVer        bool
	alterRecaptcha bool
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().IntVarP(&port, "port", "p", 0, "HTTP Port (default is random)")
	serveCmd.PersistentFlags().StringVarP(&addr, "address", "a", "", "Address to listen to (default is all of them)")
	serveCmd.PersistentFlags().StringVarP(&database, "database", "d", "./filebrowser.db", "Database file")
	serveCmd.PersistentFlags().StringVarP(&logfile, "log", "l", "stdout", "Errors logger; can use 'stdout', 'stderr' or file")
	serveCmd.PersistentFlags().StringVarP(&scope, "scope", "s", ".", "Default scope option for new users")
	serveCmd.PersistentFlags().StringVarP(&baseurl, "baseurl", "b", "", "Base URL")
	serveCmd.PersistentFlags().StringVar(&commands, "commands", "git svn hg", "Default commands option for new users")
	serveCmd.PersistentFlags().StringVar(&prefixurl, "prefixurl", "", "Prefix URL")
	serveCmd.PersistentFlags().StringVar(&viewMode, "view-mode", "mosaic", "Default view mode for new users")
	serveCmd.PersistentFlags().StringVar(&recaptchakey, "recaptcha-key", "", "ReCaptcha site key")
	serveCmd.PersistentFlags().StringVar(&recaptchasecret, "recaptcha-secret", "", "ReCaptcha secret")
	serveCmd.PersistentFlags().BoolVar(&allowCommands, "allow-commands", true, "Default allow commands option for new users")
	serveCmd.PersistentFlags().BoolVar(&allowEdit, "allow-edit", true, "Default allow edit option for new users")
	serveCmd.PersistentFlags().BoolVar(&allowPublish, "allow-publish", true, "Default allow publish option for new users")
	serveCmd.PersistentFlags().StringVar(&auth.method, "auth.method", "default", "Switch between 'none', 'default' and 'proxy' authentication.")
	serveCmd.PersistentFlags().StringVar(&auth.loginHeader, "auth.loginHeader", "X-Forwarded-User", "The header name used for proxy authentication.")
	serveCmd.PersistentFlags().BoolVar(&allowNew, "allow-new", true, "Default allow new option for new users")
	serveCmd.PersistentFlags().BoolVar(&noAuth, "no-auth", false, "Disables authentication")
	serveCmd.PersistentFlags().BoolVar(&alterRecaptcha, "alternative-recaptcha", false, "Use recaptcha.net for serving and handling, useful in China")
	serveCmd.PersistentFlags().StringVar(&locale, "locale", "", "Default locale for new users, set it empty to enable auto detect from browser")
	serveCmd.PersistentFlags().StringVar(&staticg, "staticgen", "", "Static Generator you want to enable")
	serveCmd.PersistentFlags().BoolVarP(&showVer, "version", "v", false, "Show version")

	viper.SetDefault("Address", "")
	viper.SetDefault("Port", "0")
	viper.SetDefault("Database", "./filebrowser.db")
	viper.SetDefault("Scope", ".")
	viper.SetDefault("Logger", "stdout")
	viper.SetDefault("Commands", []string{"git", "svn", "hg"})
	viper.SetDefault("AllowCommmands", true)
	viper.SetDefault("AllowEdit", true)
	viper.SetDefault("AllowNew", true)
	viper.SetDefault("AllowPublish", true)
	viper.SetDefault("StaticGen", "")
	viper.SetDefault("Locale", "")
	viper.SetDefault("AuthMethod", "default")
	viper.SetDefault("LoginHeader", "X-Fowarded-User")
	viper.SetDefault("NoAuth", false)
	viper.SetDefault("BaseURL", "")
	viper.SetDefault("PrefixURL", "")
	viper.SetDefault("ViewMode", filebrowser.MosaicViewMode)
	viper.SetDefault("AlternativeRecaptcha", false)
	viper.SetDefault("ReCaptchaKey", "")
	viper.SetDefault("ReCaptchaSecret", "")

	viper.BindPFlag("Port", serveCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("Address", serveCmd.PersistentFlags().Lookup("address"))
	viper.BindPFlag("Database", serveCmd.PersistentFlags().Lookup("database"))
	viper.BindPFlag("Scope", serveCmd.PersistentFlags().Lookup("scope"))
	viper.BindPFlag("Logger", serveCmd.PersistentFlags().Lookup("log"))
	viper.BindPFlag("Commands", serveCmd.PersistentFlags().Lookup("commands"))
	viper.BindPFlag("AllowCommands", serveCmd.PersistentFlags().Lookup("allow-commands"))
	viper.BindPFlag("AllowEdit", serveCmd.PersistentFlags().Lookup("allow-edit"))
	viper.BindPFlag("AllowNew", serveCmd.PersistentFlags().Lookup("allow-new"))
	viper.BindPFlag("AllowPublish", serveCmd.PersistentFlags().Lookup("allow-publish"))
	viper.BindPFlag("Locale", serveCmd.PersistentFlags().Lookup("locale"))
	viper.BindPFlag("StaticGen", serveCmd.PersistentFlags().Lookup("staticgen"))
	viper.BindPFlag("AuthMethod", serveCmd.PersistentFlags().Lookup("auth.method"))
	viper.BindPFlag("LoginHeader", serveCmd.PersistentFlags().Lookup("auth.loginHeader"))
	viper.BindPFlag("NoAuth", serveCmd.PersistentFlags().Lookup("no-auth"))
	viper.BindPFlag("BaseURL", serveCmd.PersistentFlags().Lookup("baseurl"))
	viper.BindPFlag("PrefixURL", serveCmd.PersistentFlags().Lookup("prefixurl"))
	viper.BindPFlag("ViewMode", serveCmd.PersistentFlags().Lookup("view-mode"))
	viper.BindPFlag("AlternativeRecaptcha", serveCmd.PersistentFlags().Lookup("alternative-recaptcha"))
	viper.BindPFlag("ReCaptchaKey", serveCmd.PersistentFlags().Lookup("recaptcha-key"))
	viper.BindPFlag("ReCaptchaSecret", serveCmd.PersistentFlags().Lookup("recaptcha-secret"))
}

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser"
	"github.com/filebrowser/filebrowser/bolt"
	h "github.com/filebrowser/filebrowser/http"
	"github.com/filebrowser/filebrowser/staticgen"
	"github.com/hacdias/fileutils"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	addr            string
	config          string
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
		method string
		header string
	}
	allowCommands  bool
	allowEdit      bool
	allowNew       bool
	allowPublish   bool
	showVer        bool
	alterRecaptcha bool
)

func init() {
	flag.StringVarP(&config, "config", "c", "", "Configuration file")
	flag.IntVarP(&port, "port", "p", 0, "HTTP Port (default is random)")
	flag.StringVarP(&addr, "address", "a", "", "Address to listen to (default is all of them)")
	flag.StringVarP(&database, "database", "d", "./filebrowser.db", "Database file")
	flag.StringVarP(&logfile, "log", "l", "stdout", "Errors logger; can use 'stdout', 'stderr' or file")
	flag.StringVarP(&baseurl, "baseurl", "b", "", "Base URL")
	flag.StringVar(&prefixurl, "prefixurl", "", "Prefix URL")
	flag.StringVar(&staticg, "staticgen", "", "Static Generator you want to enable")
	flag.BoolVarP(&showVer, "version", "v", false, "Show version")

	// User default values
	flag.StringVar(&commands, "defaults.commands", "git svn hg", "Default commands option for new users")
	flag.StringVarP(&scope, "defaults.scope", "s", ".", "Default scope option for new users")
	flag.StringVar(&viewMode, "defaults.viewMode", "mosaic", "Default view mode for new users")
	flag.BoolVar(&allowCommands, "defaults.allowCommands", true, "Default allow commands option for new users")
	flag.BoolVar(&allowEdit, "defaults.allowEdit", true, "Default allow edit option for new users")
	flag.BoolVar(&allowPublish, "defaults.allowPublish", true, "Default allow publish option for new users")
	flag.BoolVar(&allowNew, "defaults.allowNew", true, "Default allow new option for new users")
	flag.StringVar(&locale, "defaults.locale", "", "Default locale for new users, set it empty to enable auto detect from browser")

	// Recaptcha settings
	flag.BoolVar(&alterRecaptcha, "recaptcha.alternative", false, "Use recaptcha.net for serving and handling, useful in China")
	flag.StringVar(&recaptchakey, "recaptcha.key", "", "ReCaptcha site key")
	flag.StringVar(&recaptchasecret, "recaptcha.secret", "", "ReCaptcha secret")

	// Auth settings
	flag.StringVar(&auth.method, "auth.method", "default", "Switch between 'none', 'default' and 'proxy' authentication")
	flag.StringVar(&auth.header, "auth.header", "X-Forwarded-User", "The header name used for proxy authentication")
}

func setupViper() {
	viper.SetDefault("Port", "0")
	viper.SetDefault("Address", "")
	viper.SetDefault("Database", "./filebrowser.db")
	viper.SetDefault("Logger", "stdout")
	viper.SetDefault("BaseURL", "")
	viper.SetDefault("PrefixURL", "")
	viper.SetDefault("StaticGen", "")

	viper.BindPFlag("Port", flag.Lookup("port"))
	viper.BindPFlag("Address", flag.Lookup("address"))
	viper.BindPFlag("Database", flag.Lookup("database"))
	viper.BindPFlag("Logger", flag.Lookup("log"))
	viper.BindPFlag("BaseURL", flag.Lookup("baseurl"))
	viper.BindPFlag("PrefixURL", flag.Lookup("prefixurl"))
	viper.BindPFlag("StaticGen", flag.Lookup("staticgen"))

	// User default values
	viper.SetDefault("Defaults.Scope", ".")
	viper.SetDefault("Defaults.Commands", []string{"git", "svn", "hg"})
	viper.SetDefault("Defaults.ViewMode", filebrowser.MosaicViewMode)
	viper.SetDefault("Defaults.AllowCommmands", true)
	viper.SetDefault("Defaults.AllowEdit", true)
	viper.SetDefault("Defaults.AllowNew", true)
	viper.SetDefault("Defaults.AllowPublish", true)
	viper.SetDefault("Defaults.Locale", "")

	viper.BindPFlag("Defaults.Scope", flag.Lookup("defaults.scope"))
	viper.BindPFlag("Defaults.Commands", flag.Lookup("defaults.commands"))
	viper.BindPFlag("Defaults.ViewMode", flag.Lookup("defaults.viewMode"))
	viper.BindPFlag("Defaults.AllowCommands", flag.Lookup("defaults.allowCommands"))
	viper.BindPFlag("Defaults.AllowEdit", flag.Lookup("defaults.allowEdit"))
	viper.BindPFlag("Defaults.AllowNew", flag.Lookup("defaults.allowNew"))
	viper.BindPFlag("Defaults.AllowPublish", flag.Lookup("defaults.allowPublish"))
	viper.BindPFlag("Defaults.Locale", flag.Lookup("defaults.locale"))

	// Recaptcha settings
	viper.SetDefault("Recaptcha.Alternative", false)
	viper.SetDefault("Recaptcha.Key", "")
	viper.SetDefault("Recaptcha.Secret", "")

	viper.BindPFlag("Recaptcha.Alternative", flag.Lookup("recaptcha.alternative"))
	viper.BindPFlag("Recaptcha.Key", flag.Lookup("recaptcha.key"))
	viper.BindPFlag("Recaptcha.Secret", flag.Lookup("recaptcha.secret"))

	// Auth settings
	viper.SetDefault("Auth.Method", "default")
	viper.SetDefault("Auth.Header", "X-Fowarded-User")

	viper.BindPFlag("Auth.Method", flag.Lookup("auth.method"))
	viper.BindPFlag("Auth.Header", flag.Lookup("auth.header"))

	viper.SetConfigName("filebrowser")
	viper.AddConfigPath(".")
}

func printVersion() {
	fmt.Println("filebrowser version", filebrowser.Version)
	os.Exit(0)
}

func main() {
	setupViper()
	flag.Parse()

	if showVer {
		printVersion()
	}

	// Add a configuration file if set.
	if config != "" {
		ext := filepath.Ext(config)
		dir := filepath.Dir(config)
		config = strings.TrimSuffix(config, ext)

		if dir != "" {
			viper.AddConfigPath(dir)
			config = strings.TrimPrefix(config, dir)
		}

		viper.SetConfigName(config)
	}

	// Read configuration from a file if exists.
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigParseError); ok {
			panic(err)
		}
	}

	// Set up process log before anything bad happens.
	switch viper.GetString("Logger") {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "":
		log.SetOutput(ioutil.Discard)
	default:
		log.SetOutput(&lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		})
	}

	// Validate the provided config before moving forward
	if viper.GetString("Auth.Method") != "none" && viper.GetString("Auth.Method") != "default" && viper.GetString("Auth.Method") != "proxy" {
		log.Fatal("The property 'auth.method' needs to be set to 'default' or 'proxy'.")
	}

	if viper.GetString("Auth.Method") == "proxy" {
		if viper.GetString("Auth.Header") == "" {
			log.Fatal("The 'auth.header' needs to be specified when 'proxy' authentication is used.")
		}
		log.Println("[WARN] Filebrowser authentication is configured to 'proxy' authentication. This can cause a huge security issue if the infrastructure is not configured correctly.")
	}

	// Builds the address and a listener.
	laddr := viper.GetString("Address") + ":" + viper.GetString("Port")
	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}

	// Tell the user the port in which is listening.
	fmt.Println("Listening on", listener.Addr().String())

	// Starts the server.
	if err := http.Serve(listener, handler()); err != nil {
		log.Fatal(err)
	}
}

func handler() http.Handler {
	db, err := storm.Open(viper.GetString("Database"))
	if err != nil {
		log.Fatal(err)
	}

	recaptchaHost := "https://www.google.com"
	if viper.GetBool("Recaptcha.Alternative") {
		recaptchaHost = "https://recaptcha.net"
	}

	fm := &filebrowser.FileBrowser{
		Auth: &filebrowser.Auth{
			Method: viper.GetString("Auth.Method"),
			Header: viper.GetString("Auth.Header"),
		},
		ReCaptcha: &filebrowser.ReCaptcha{
			Host:   recaptchaHost,
			Key:    viper.GetString("Recaptcha.Key"),
			Secret: viper.GetString("Recaptcha.Secret"),
		},
		DefaultUser: &filebrowser.User{
			AllowCommands: viper.GetBool("Defaults.AllowCommands"),
			AllowEdit:     viper.GetBool("Defaults.AllowEdit"),
			AllowNew:      viper.GetBool("Defaults.AllowNew"),
			AllowPublish:  viper.GetBool("Defaults.AllowPublish"),
			Commands:      viper.GetStringSlice("Defaults.Commands"),
			Rules:         []*filebrowser.Rule{},
			Locale:        viper.GetString("Defaults.Locale"),
			CSS:           "",
			Scope:         viper.GetString("Defaults.Scope"),
			FileSystem:    fileutils.Dir(viper.GetString("Defaults.Scope")),
			ViewMode:      viper.GetString("Defaults.ViewMode"),
		},
		Store: &filebrowser.Store{
			Config: bolt.ConfigStore{DB: db},
			Users:  bolt.UsersStore{DB: db},
			Share:  bolt.ShareStore{DB: db},
		},
		NewFS: func(scope string) filebrowser.FileSystem {
			return fileutils.Dir(scope)
		},
	}

	fm.SetBaseURL(viper.GetString("BaseURL"))
	fm.SetPrefixURL(viper.GetString("PrefixURL"))

	err = fm.Setup()
	if err != nil {
		log.Fatal(err)
	}

	switch viper.GetString("StaticGen") {
	case "hugo":
		hugo := &staticgen.Hugo{
			Root:        viper.GetString("Scope"),
			Public:      filepath.Join(viper.GetString("Scope"), "public"),
			Args:        []string{},
			CleanPublic: true,
		}

		if err = fm.Attach(hugo); err != nil {
			log.Fatal(err)
		}
	case "jekyll":
		jekyll := &staticgen.Jekyll{
			Root:        viper.GetString("Scope"),
			Public:      filepath.Join(viper.GetString("Scope"), "_site"),
			Args:        []string{"build"},
			CleanPublic: true,
		}

		if err = fm.Attach(jekyll); err != nil {
			log.Fatal(err)
		}
	}

	return h.Handler(fm)
}

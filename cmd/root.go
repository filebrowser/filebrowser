package cmd

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/filebrowser/filebrowser/v2/auth"
	fbhttp "github.com/filebrowser/filebrowser/v2/http"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/filebrowser/filebrowser/v2/version"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v "github.com/spf13/viper"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	cfgFile string
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.SetVersionTemplate("File Browser version {{printf \"%s\" .Version}}\n")

	flags := rootCmd.Flags()
	persistent := rootCmd.PersistentFlags()

	persistent.StringVarP(&cfgFile, "config", "c", "", "config file path")
	persistent.StringP("database", "d", "./filebrowser.db", "database path")
	flags.String("username", "admin", "username for the first user when using quick config")
	flags.String("password", "", "hashed password for the first user when using quick config (default \"admin\")")

	addServerFlags(flags)
}

func addServerFlags(flags *pflag.FlagSet) {
	flags.StringP("address", "a", "127.0.0.1", "address to listen on")
	flags.StringP("log", "l", "stdout", "log output")
	flags.StringP("port", "p", "8080", "port to listen on")
	flags.StringP("cert", "t", "", "tls certificate")
	flags.StringP("key", "k", "", "tls key")
	flags.StringP("root", "r", ".", "root to prepend to relative paths")
	flags.StringP("baseurl", "b", "", "base url")
}

// NOTE: we could simply bind the flags to viper and use IsSet.
// Although there is a bug on Viper that always returns true on IsSet
// if a flag is binded. Our alternative way is to manually check
// the flag and then the value from env/config/gotten by viper.
// https://github.com/spf13/viper/pull/331
func getStringViperFlag(flags *pflag.FlagSet, key string) (string, bool) {
	value := ""
	set := false

	// If set on Flags, use it.
	flags.Visit(func(flag *pflag.Flag) {
		if flag.Name == key {
			set = true
			value, _ = flags.GetString(key)
		}
	})

	if set {
		return value, set
	}

	// If set through viper (env, config), return it.
	if v.IsSet(key) {
		return v.GetString(key), true
	}

	// Otherwise use default value on flags.
	value, _ = flags.GetString(key)
	return value, false
}

func mustGetStringViperFlag(flags *pflag.FlagSet, key string) string {
	val, _ := getStringViperFlag(flags, key)
	return val
}

var rootCmd = &cobra.Command{
	Use:     "filebrowser",
	Version: version.Version,
	Short:   "A stylish web-based file browser",
	Long: `File Browser CLI lets you create the database to use with File Browser,
manage your users and all the configurations without acessing the
web interface.

If you've never run File Browser, you'll need to have a database for
it. Don't worry: you don't need to setup a separate database server.
We're using Bolt DB which is a single file database and all managed
by ourselves.

For this specific command, all the flags you have available (except
"config" for the configuration file), can be given either through
environment variables or configuration files.

If you don't set "config", it will look for a configuration file called
.filebrowser.{json, toml, yaml, yml} in the following directories:

- ./
- $HOME/
- /etc/filebrowser/

The precedence of the configuration values are as follows:

- flags
- environment variables
- configuration file
- database values
- defaults

The environment variables are prefixed by "FB_" followed by the option
name in caps. So to set "database" via an env variable, you should
set FB_DATABASE.

Also, if the database path doesn't exist, File Browser will enter into
the quick setup mode and a new database will be bootstraped and a new
user created with the credentials from options "username" and "password".`,
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		log.Println(cfgFile)

		if !d.hadDB {
			quickSetup(cmd.Flags(), d)
		}

		server := getServerWithViper(cmd.Flags(), d.store)
		setupLog(server.Log)

		root, err := filepath.Abs(server.Root)
		checkErr(err)
		server.Root = root

		adr := server.Address + ":" + server.Port

		var listener net.Listener

		if server.TLSKey != "" && server.TLSCert != "" {
			cer, err := tls.LoadX509KeyPair(server.TLSCert, server.TLSKey)
			checkErr(err)
			listener, err = tls.Listen("tcp", adr, &tls.Config{Certificates: []tls.Certificate{cer}})
			checkErr(err)
		} else {
			listener, err = net.Listen("tcp", adr)
			checkErr(err)
		}

		handler, err := fbhttp.NewHandler(d.store, server)
		checkErr(err)

		log.Println("Listening on", listener.Addr().String())
		if err := http.Serve(listener, handler); err != nil {
			log.Fatal(err)
		}
	}, pythonConfig{allowNoDB: true}),
}

func getServerWithViper(flags *pflag.FlagSet, st *storage.Storage) *settings.Server {
	server, err := st.Settings.GetServer()
	checkErr(err)

	if val, set := getStringViperFlag(flags, "root"); set {
		server.Root = val
	}

	if val, set := getStringViperFlag(flags, "baseurl"); set {
		server.BaseURL = val
	}

	if val, set := getStringViperFlag(flags, "address"); set {
		server.Address = val
	}

	if val, set := getStringViperFlag(flags, "port"); set {
		server.Port = val
	}

	if val, set := getStringViperFlag(flags, "log"); set {
		server.Log = val
	}

	if val, set := getStringViperFlag(flags, "key"); set {
		server.TLSKey = val
	}

	if val, set := getStringViperFlag(flags, "cert"); set {
		server.TLSCert = val
	}

	return server
}

func setupLog(logMethod string) {
	switch logMethod {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "":
		log.SetOutput(ioutil.Discard)
	default:
		log.SetOutput(&lumberjack.Logger{
			Filename:   logMethod,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		})
	}

}

func quickSetup(flags *pflag.FlagSet, d pythonData) {
	set := &settings.Settings{
		Key:        generateRandomBytes(64), // 256 bit
		Signup:     false,
		AuthMethod: auth.MethodJSONAuth,
		Defaults: settings.UserDefaults{
			Scope:  ".",
			Locale: "en",
			Perm: users.Permissions{
				Admin:    false,
				Execute:  true,
				Create:   true,
				Rename:   true,
				Modify:   true,
				Delete:   true,
				Share:    true,
				Download: true,
			},
		},
	}

	ser := &settings.Server{
		BaseURL: mustGetStringViperFlag(flags, "baseurl"),
		Port:    mustGetStringViperFlag(flags, "port"),
		Log:     mustGetStringViperFlag(flags, "log"),
		TLSKey:  mustGetStringViperFlag(flags, "key"),
		TLSCert: mustGetStringViperFlag(flags, "cert"),
		Address: mustGetStringViperFlag(flags, "address"),
		Root:    mustGetStringViperFlag(flags, "root"),
	}

	err := d.store.Settings.Save(set)
	checkErr(err)

	err = d.store.Settings.SaveServer(ser)
	checkErr(err)

	err = d.store.Auth.Save(&auth.JSONAuth{})
	checkErr(err)

	username := mustGetStringViperFlag(flags, "username")
	password := mustGetStringViperFlag(flags, "password")

	if password == "" {
		password, err = users.HashPwd("admin")
		checkErr(err)
	}

	if username == "" || password == "" {
		log.Fatal("username and password cannot be empty during quick setup")
	}

	user := &users.User{
		Username:     username,
		Password:     password,
		LockPassword: false,
	}

	set.Defaults.Apply(user)
	user.Perm.Admin = true

	err = d.store.Users.Save(user)
	checkErr(err)
}

func initConfig() {
	if cfgFile == "" {
		home, err := homedir.Dir()
		checkErr(err)
		v.AddConfigPath(".")
		v.AddConfigPath(home)
		v.AddConfigPath("/etc/filebrowser/")
		v.SetConfigName(".filebrowser")
	} else {
		v.SetConfigFile(cfgFile)
	}

	v.SetEnvPrefix("FB")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(v.ConfigParseError); ok {
			panic(err)
		}
		cfgFile = "No config file used"
	} else {
		cfgFile = "Using config file: " + v.ConfigFileUsed()
	}

}

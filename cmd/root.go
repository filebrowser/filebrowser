package cmd

import (
	"crypto/tls"
	"errors"
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
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	v "github.com/spf13/viper"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	cfgFile string
)

func init() {
	cobra.OnInitialize(initConfig)

	f := rootCmd.Flags()
	pf := rootCmd.PersistentFlags()

	pf.StringVarP(&cfgFile, "config", "c", "", "config file path")
	vaddP(pf, "database", "d", "./filebrowser.db", "path to the database")
	vaddP(f, "address", "a", "127.0.0.1", "address to listen on")
	vaddP(f, "log", "l", "stdout", "log output")
	vaddP(f, "port", "p", "8080", "port to listen on")
	vaddP(f, "cert", "t", "", "tls certificate")
	vaddP(f, "key", "k", "", "tls key")
	vaddP(f, "root", "r", ".", "root to prepend to relative paths")
	vaddP(f, "baseurl", "b", "", "base url")
	vadd(f, "username", "admin", "username for the first user when using quick config")
	vadd(f, "password", "", "hashed password for the first user when using quick config (default \"admin\")")

	if err := v.BindPFlags(f); err != nil {
		panic(err)
	}

	if err := v.BindPFlags(pf); err != nil {
		panic(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "filebrowser",
	Short: "A stylish web-based file browser",
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

- flag
- environment variable
- configuration file
- defaults

The environment variables are prefixed by "FB_" followed by the option
name in caps. So to set "database" via an env variable, you should
set FB_DATABASE equals to the path.

Also, if the database path doesn't exist, File Browser will enter into
the quick setup mode and a new database will be bootstraped and a new
user created with the credentials from options "username" and "password".`,
	Run: python(func(cmd *cobra.Command, args []string, d pythonData) {
		if !d.hadDB {
			quickSetup(d)
		}

		server := getServer(d.store)
		setupLog(server.Log)

		handler, err := fbhttp.NewHandler(d.store, server)
		checkErr(err)

		var listener net.Listener

		if server.TLSKey != "" && server.TLSCert != "" {
			cer, err := tls.LoadX509KeyPair(server.TLSCert, server.TLSKey)
			checkErr(err)
			config := &tls.Config{Certificates: []tls.Certificate{cer}}
			listener, err = tls.Listen("tcp", server.Address+":"+server.Port, config)
			checkErr(err)
		} else {
			listener, err = net.Listen("tcp", server.Address+":"+server.Port)
			checkErr(err)
		}

		log.Println("Listening on", listener.Addr().String())
		if err := http.Serve(listener, handler); err != nil {
			log.Fatal(err)
		}
	}, pythonConfig{allowNoDB: true}),
}

// TODO: get server settings and only replace
// them if set on Viper. Although viper.IsSet
// is bugged and if binded to a pflag, it will
// always return true.
// Also, when doing that, add this options to
// config init, config import, printConfig
// and config set since the DB values will actually
// be used. For now, despite being stored in the DB,
// they won't be used.
func getServer(st *storage.Storage) *settings.Server {
	root := v.GetString("root")
	root, err := filepath.Abs(root)
	checkErr(err)
	server := &settings.Server{}
	server.BaseURL = v.GetString("baseurl")
	server.Root = root
	server.Address = v.GetString("address")
	server.Port = v.GetString("port")
	server.TLSKey = v.GetString("key")
	server.TLSCert = v.GetString("cert")
	server.Log = v.GetString("log")
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

func quickSetup(d pythonData) {
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
		BaseURL: v.GetString("baseurl"),
		Log:     v.GetString("log"),
		TLSKey:  v.GetString("key"),
		TLSCert: v.GetString("cert"),
		Address: v.GetString("address"),
		Root:    v.GetString("root"),
	}

	err := d.store.Settings.Save(set)
	checkErr(err)

	err = d.store.Settings.SaveServer(ser)
	checkErr(err)

	err = d.store.Auth.Save(&auth.JSONAuth{})
	checkErr(err)

	username := v.GetString("username")
	password := v.GetString("password")

	if password == "" {
		password, err = users.HashPwd("admin")
		checkErr(err)
	}

	if username == "" || password == "" {
		checkErr(errors.New("username and password cannot be empty during quick setup"))
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
		// TODO: log.Println("No config file provided")
	}
	// else TODO: log.Println("Using config file:", v.ConfigFileUsed())
}

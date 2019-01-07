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
	"strconv"
	"strings"

	"github.com/filebrowser/filebrowser/v2/auth"
	fbhttp "github.com/filebrowser/filebrowser/v2/http"
	"github.com/filebrowser/filebrowser/v2/settings"
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
	vaddP(f, "port", "p", 8080, "port to listen on")
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
		switch logMethod := v.GetString("log"); logMethod {
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

		if !d.hadDB {
			quickSetup(d)
		}

		port := v.GetInt("port")
		address := v.GetString("address")
		cert := v.GetString("cert")
		key := v.GetString("key")
		root := v.GetString("root")

		root, err := filepath.Abs(root)
		checkErr(err)
		settings, err := d.store.Settings.Get()
		checkErr(err)

		// Despite Base URL and Scope being "server" type of
		// variables, we persist them to the database because
		// they are needed during the execution and not only
		// to start up the server.
		settings.BaseURL = v.GetString("baseurl")
		settings.Root = root
		err = d.store.Settings.Save(settings)
		checkErr(err)

		handler, err := fbhttp.NewHandler(d.store)
		checkErr(err)

		var listener net.Listener

		if key != "" && cert != "" {
			cer, err := tls.LoadX509KeyPair(cert, key)
			checkErr(err)
			config := &tls.Config{Certificates: []tls.Certificate{cer}}
			listener, err = tls.Listen("tcp", address+":"+strconv.Itoa(port), config)
			checkErr(err)
		} else {
			listener, err = net.Listen("tcp", address+":"+strconv.Itoa(port))
			checkErr(err)
		}

		log.Println("Listening on", listener.Addr().String())
		if err := http.Serve(listener, handler); err != nil {
			log.Fatal(err)
		}
	}, pythonConfig{allowNoDB: true}),
}

func quickSetup(d pythonData) {
	set := &settings.Settings{
		Key:        generateRandomBytes(64), // 256 bit
		BaseURL:    v.GetString("baseurl"),
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

	err := d.store.Settings.Save(set)
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

package cmd

import (
	"crypto/tls"
	"fmt"
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
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/filebrowser/filebrowser/v2/version"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	v "github.com/spf13/viper"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	cfgFile string
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SetVersionTemplate("File Browser version {{printf \"%s\" .Version}}\n")

	f := rootCmd.Flags()
	pf := rootCmd.PersistentFlags()

	pf.StringVarP(&cfgFile, "config", "c", "", "config file path")
	vaddP(pf, "database", "d", "./filebrowser.db", "path to the database")
	vaddP(f, "address", "a", settings.RuntimeDefaults["address"], "address to listen on")
	vaddP(f, "log", "l", settings.RuntimeDefaults["log"], "log output")
	vaddP(f, "port", "p", settings.RuntimeDefaults["port"], "port to listen on")
	vaddP(f, "cert", "t", settings.RuntimeDefaults["cert"], "tls certificate")
	vaddP(f, "key", "k", settings.RuntimeDefaults["key"], "tls key")
	vaddP(f, "root", "r", settings.RuntimeDefaults["root"], "root path to prepend to all relative paths")
	vaddP(f, "baseurl", "b", settings.RuntimeDefaults["baseurl"], "base url")
	vadd(f, "username", "admin", "username for the first user when using quick config")
	vadd(f, "password", "", "hashed password for the first user when using quick config (default \"admin\")")
	vaddP(f, "force", "f", false, "")

	if err := v.BindPFlags(f); err != nil {
		panic(err)
	}

	if err := v.BindPFlags(pf); err != nil {
		panic(err)
	}
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

		log.Println(cfgFile)

		switch logMethod := settings.RuntimeCfg["log"]; logMethod {
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

		s, err := d.store.Settings.Get()
		checkErr(err)

		if !v.GetBool("force") {
			for k := range settings.RuntimeCfg {
				if y, ok := s.Runtime[k]; ok {
					settings.RuntimeCfg[k] = y
				}
			}
		}

		r, err := filepath.Abs(settings.RuntimeCfg["root"])
		checkErr(err)
		settings.RuntimeCfg["root"] = r

		adr := settings.RuntimeCfg["address"] + ":" + settings.RuntimeCfg["port"]
		cert := settings.RuntimeCfg["cert"]
		key := settings.RuntimeCfg["key"]

		var listener net.Listener

		if key != "" && cert != "" {
			cer, err := tls.LoadX509KeyPair(cert, key)
			checkErr(err)
			listener, err = tls.Listen("tcp", adr, &tls.Config{Certificates: []tls.Certificate{cer}})
			checkErr(err)
		} else {
			listener, err = net.Listen("tcp", adr)
			checkErr(err)
		}

		handler, err := fbhttp.NewHandler(d.store)
		checkErr(err)

		log.Println("Listening on", listener.Addr().String())
		if err := http.Serve(listener, handler); err != nil {
			log.Fatal(err)
		}
	}, pythonConfig{allowNoDB: true}),
}

func runtimeNonDefaults() map[string]string {
	m := map[string]string{}
	for k, d := range settings.RuntimeDefaults {
		if x, ok := settings.RuntimeCfg[k]; ok && (x != d) {
			log.Println(fmt.Sprintf("Non-default value for key '%s': %s [default: %s]", k, x, d))
			m[k] = x
		}
	}
	if settings.RuntimeCfg["key"] == "" {
		log.Println("Generate random 256 bit key")
		m["key"] = string(generateRandomBytes(64)) // 256 bit
	}
	return m
}

func quickSetup(d pythonData) {

	log.Println("Executing quick setup...")

	set := &settings.Settings{
		Runtime:    runtimeNonDefaults(),
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

	log.Println("Quick setup finished")
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

	for k := range settings.RuntimeDefaults {
		settings.RuntimeCfg[k] = v.GetString(k)
	}
}

package cmd

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v "github.com/spf13/viper"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/diskcache"
	fbhttp "github.com/filebrowser/filebrowser/v2/http"
	"github.com/filebrowser/filebrowser/v2/img"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
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
	flags.Bool("noauth", false, "use the noauth auther when using quick setup")
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
	flags.String("socket", "", "socket to listen to (cannot be used with address, port, cert nor key flags)")
	flags.StringP("baseurl", "b", "", "base url")
	flags.String("cache-dir", "", "file cache directory (disabled if empty)")
	flags.Int("img-processors", 4, "image processors count")
	flags.Bool("disable-thumbnails", false, "disable image thumbnails")
	flags.Bool("disable-preview-resize", false, "disable resize of image previews")
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

		// build img service
		workersCount, err := cmd.Flags().GetInt("img-processors")
		checkErr(err)
		if workersCount < 1 {
			log.Fatal("Image resize workers count could not be < 1")
		}
		imgSvc := img.New(workersCount)

		var fileCache diskcache.Interface = diskcache.NewNoOp()
		cacheDir, err := cmd.Flags().GetString("cache-dir")
		checkErr(err)
		if cacheDir != "" {
			if err := os.MkdirAll(cacheDir, 0700); err != nil { //nolint:govet
				log.Fatalf("can't make directory %s: %s", cacheDir, err)
			}
			fileCache = diskcache.New(afero.NewOsFs(), cacheDir)
		}

		server := getRunParams(cmd.Flags(), d.store)
		setupLog(server.Log)

		root, err := filepath.Abs(server.Root)
		checkErr(err)
		server.Root = root

		adr := server.Address + ":" + server.Port

		var listener net.Listener

		switch {
		case server.Socket != "":
			listener, err = net.Listen("unix", server.Socket)
			checkErr(err)
		case server.TLSKey != "" && server.TLSCert != "":
			cer, err := tls.LoadX509KeyPair(server.TLSCert, server.TLSKey) //nolint:shadow
			checkErr(err)
			listener, err = tls.Listen("tcp", adr, &tls.Config{Certificates: []tls.Certificate{cer}}) //nolint:shadow
			checkErr(err)
		default:
			listener, err = net.Listen("tcp", adr) //nolint:shadow
			checkErr(err)
		}

		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
		go cleanupHandler(listener, sigc)

		handler, err := fbhttp.NewHandler(imgSvc, fileCache, d.store, server)
		checkErr(err)

		defer listener.Close()

		log.Println("Listening on", listener.Addr().String())
		if err := http.Serve(listener, handler); err != nil {
			log.Fatal(err)
		}
	}, pythonConfig{allowNoDB: true}),
}

func cleanupHandler(listener net.Listener, c chan os.Signal) { //nolint:interfacer
	sig := <-c
	log.Printf("Caught signal %s: shutting down.", sig)
	listener.Close()
	os.Exit(0)
}

//nolint:gocyclo
func getRunParams(flags *pflag.FlagSet, st *storage.Storage) *settings.Server {
	server, err := st.Settings.GetServer()
	checkErr(err)

	if val, set := getParamB(flags, "root"); set {
		server.Root = val
	}

	if val, set := getParamB(flags, "baseurl"); set {
		server.BaseURL = val
	}

	if val, set := getParamB(flags, "log"); set {
		server.Log = val
	}

	isSocketSet := false
	isAddrSet := false

	if val, set := getParamB(flags, "address"); set {
		server.Address = val
		isAddrSet = isAddrSet || set
	}

	if val, set := getParamB(flags, "port"); set {
		server.Port = val
		isAddrSet = isAddrSet || set
	}

	if val, set := getParamB(flags, "key"); set {
		server.TLSKey = val
		isAddrSet = isAddrSet || set
	}

	if val, set := getParamB(flags, "cert"); set {
		server.TLSCert = val
		isAddrSet = isAddrSet || set
	}

	if val, set := getParamB(flags, "socket"); set {
		server.Socket = val
		isSocketSet = isSocketSet || set
	}

	if isAddrSet && isSocketSet {
		checkErr(errors.New("--socket flag cannot be used with --address, --port, --key nor --cert"))
	}

	// Do not use saved Socket if address was manually set.
	if isAddrSet && server.Socket != "" {
		server.Socket = ""
	}

	_, disableThumbnails := getParamB(flags, "disable-thumbnails")
	server.EnableThumbnails = !disableThumbnails

	_, disablePreviewResize := getParamB(flags, "disable-preview-resize")
	server.ResizePreview = !disablePreviewResize

	return server
}

// getParamB returns a parameter as a string and a boolean to tell if it is different from the default
//
// NOTE: we could simply bind the flags to viper and use IsSet.
// Although there is a bug on Viper that always returns true on IsSet
// if a flag is binded. Our alternative way is to manually check
// the flag and then the value from env/config/gotten by viper.
// https://github.com/spf13/viper/pull/331
func getParamB(flags *pflag.FlagSet, key string) (string, bool) {
	value, _ := flags.GetString(key)

	// If set on Flags, use it.
	if flags.Changed(key) {
		return value, true
	}

	// If set through viper (env, config), return it.
	if v.IsSet(key) {
		return v.GetString(key), true
	}

	// Otherwise use default value on flags.
	return value, false
}

func getParam(flags *pflag.FlagSet, key string) string {
	val, _ := getParamB(flags, key)
	return val
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
		Key:           generateKey(),
		Signup:        false,
		CreateUserDir: false,
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

	var err error
	if _, noauth := getParamB(flags, "noauth"); noauth {
		set.AuthMethod = auth.MethodNoAuth
		err = d.store.Auth.Save(&auth.NoAuth{})
	} else {
		set.AuthMethod = auth.MethodJSONAuth
		err = d.store.Auth.Save(&auth.JSONAuth{})
	}

	checkErr(err)
	err = d.store.Settings.Save(set)
	checkErr(err)

	ser := &settings.Server{
		BaseURL: getParam(flags, "baseurl"),
		Port:    getParam(flags, "port"),
		Log:     getParam(flags, "log"),
		TLSKey:  getParam(flags, "key"),
		TLSCert: getParam(flags, "cert"),
		Address: getParam(flags, "address"),
		Root:    getParam(flags, "root"),
	}

	err = d.store.Settings.SaveServer(ser)
	checkErr(err)

	username := getParam(flags, "username")
	password := getParam(flags, "password")

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

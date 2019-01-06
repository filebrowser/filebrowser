package cmd

import (
	"crypto/rand"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"

	fbhttp "github.com/filebrowser/filebrowser/v2/http"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	//	"github.com/spf13/pflag"
	v "github.com/spf13/viper"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var rootCmd = &cobra.Command{
	Use:   "filebrowser",
	Short: "A stylish web-based file browser",
	Long: `File Browser CLI lets you create the database to use with File Browser,
manage your user and all the configurations without accessing the
web interface.

If you've never run File Browser, you will need to create the database.
See 'filebrowser help config init' for more information.

This command is used to start up the server. By default it starts listening
on localhost on a random port unless specified otherwise in the database or
via flags.

Use the available flags to override the database/default options. These flags
values won't be persisted to the database. To persist configuration to the database
use the command 'filebrowser config set'.`,
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		defer db.Close()
		st := getStorage(db)
		startServer(st)
	},
}

var (
	cfgFile string
)

// POSSIBLE WORKAROUND TO IDENTIFY WHEN DEFAULT VALUES ARE BEING USED
var defaults = struct {
	database string
	address  string
	log      string
	port     int
	scope    string
	admin    string
}{
	"./filebrowser.db",
	"127.0.0.1",
	"stderr",
	80,
	"/srv",
	"admin",
}

func init() {
	cobra.OnInitialize(initConfig)
	//rootCmd.SetVersionTemplate("File Browser {{printf \"version %s\" .Version}}\n")

	f := rootCmd.Flags()
	pf := rootCmd.PersistentFlags()

	pf.StringVarP(&cfgFile, "config", "c", "", "config file (defaults are './.filebrowser[ext]', '$HOME/.filebrowser[ext]' or '/etc/filebrowser/.filebrowser[ext]')")

	vaddP(pf, "database", "d", "./filebrowser.db", "path to the database")

	vaddP(f, "address", "a", defaults.address, "address to listen on")
	vaddP(f, "log", "l", defaults.log, "log output")
	vaddP(f, "port", "p", defaults.port, "port to listen on")
	vaddP(f, "cert", "t", "", "tls certificate (default comes from database)")
	vaddP(f, "key", "k", "", "tls key (default comes from database)")
	vaddP(f, "scope", "s", defaults.scope, "scope for users")
	vaddP(f, "force", "f", false, "overwrite DB config with runtime params")
	vaddP(f, "admin", "f", defaults.admin, "first username")
	vaddP(f, "passwd", "f", "", "first username password hash")
	vaddP(f, "baseurl", "b", "", "base URL")

	// Bind the full flag sets to the configuration
	if err := v.BindPFlags(f); err != nil {
		panic(err)
	}
	if err := v.BindPFlags(pf); err != nil {
		panic(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile == "" {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}
		v.AddConfigPath(".")
		v.AddConfigPath(home)
		v.AddConfigPath("/etc/filebrowser/")
		v.SetConfigName(".filebrowser")
	} else {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	}

	v.SetEnvPrefix("FB")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(v.ConfigParseError); ok {
			panic(err)
		}
		log.Println("No config file provided")
	} else {
		log.Println("Using config file:", v.ConfigFileUsed())
	}

	log.Println("FORCE:", v.GetBool("force"))

	/*
	   if DB exists
	     if force false
	       database has highest priority, if undefined in DB use config params
	     else
	       config params overwrite existing and non-existing params in DB
	   else
	     (quick)Setup with provided config params
	*/

	/*
	   DISPLAY WARNINGS WHEN DEFAULT VALUES ARE USED

	     This allows to know if a CLI flag was provided:

	       log.Println(rootCmd.Flags().Changed("database"))

	     However, that is not enough in order to know if a value came from a config file or from envvars.
	     This should allow so. But it seems not to work as expected (see spf13/viper#323):

	       log.Println(v.IsSet("database"))
	*/

	if _, err := os.Stat(v.GetString("database")); os.IsNotExist(err) {
		quickSetup()
	}

}

/*
func serverVisitAndReplace(s *settings.Settings) {
	rootCmd.Flags().Visit(func(flag *pflag.Flag) {
		switch flag.Name {
		case "log":
			s.Log = v.GetString(flag.Name)
		case "address":
			s.Server.Address = v.GetString(flag.Name)
		case "port":
			s.Server.Port = v.GetInt(flag.Name)
		case "cert":
			s.Server.TLSCert = v.GetString(flag.Name)
		case "key":
			s.Server.TLSKey = v.GetString(flag.Name)
		}
	})
}
*/

func quickSetup() {
	scope := v.GetString("scope")
	if scope == defaults.scope {
		log.Println("[WARN] Using default value '/srv' as param 'scope'")
	}

	db, err := storm.Open(v.GetString("database"))
	checkErr(err)
	defer db.Close()

	set := &settings.Settings{
		Key:        generateRandomBytes(64), // 256 bit
		BaseURL:    v.GetString("baseurl"),
		Log:        v.GetString("log"),
		Signup:     false,
		AuthMethod: auth.MethodJSONAuth,
		Server: settings.Server{
			Port:    v.GetInt("port"),
			Address: v.GetString("address"),
			TLSCert: v.GetString("cert"),
			TLSKey:  v.GetString("key"),
		},
		Defaults: settings.UserDefaults{
			Scope:  scope,
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

	//	serverVisitAndReplace(set)
	st := getStorage(db)

	err = st.Settings.Save(set)
	checkErr(err)

	err = st.Auth.Save(&auth.JSONAuth{})
	checkErr(err)

	password := v.GetString("password")
	if password == "" {
		password, err = users.HashPwd("admin")
		checkErr(err)
	}

	user := &users.User{
		Username:     v.GetString("admin"),
		Password:     password,
		LockPassword: false,
	}

	set.Defaults.Apply(user)
	user.Perm.Admin = true

	err = st.Users.Save(user)
	checkErr(err)
}

func setupLogger(s *settings.Settings) {
	switch s.Log {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "":
		log.SetOutput(ioutil.Discard)
	default:
		log.SetOutput(&lumberjack.Logger{
			Filename:   s.Log,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		})
	}
}

func startServer(st *storage.Storage) {
	settings, err := st.Settings.Get()
	checkErr(err)

	//	serverVisitAndReplace(settings)
	setupLogger(settings)

	handler, err := fbhttp.NewHandler(st)
	checkErr(err)

	var listener net.Listener

	if settings.Server.TLSKey != "" && settings.Server.TLSCert != "" {
		cer, err := tls.LoadX509KeyPair(settings.Server.TLSCert, settings.Server.TLSKey)
		checkErr(err)
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		listener, err = tls.Listen("tcp", settings.Server.Address+":"+strconv.Itoa(settings.Server.Port), config)
		checkErr(err)
	} else {
		listener, err = net.Listen("tcp", settings.Server.Address+":"+strconv.Itoa(settings.Server.Port))
		checkErr(err)
	}

	log.Println("Listening on", listener.Addr().String())
	if err := http.Serve(listener, handler); err != nil {
		log.Fatal(err)
	}
}

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	checkErr(err)
	// Note that err == nil only if we read len(b) bytes.
	return b
}

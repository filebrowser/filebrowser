package cmd

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/asdine/storm"
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
	f := rootCmd.Flags()
	pf := rootCmd.PersistentFlags()

	f.StringVarP(&cfgFile, "config", "c", "", "config file (defaults are './.filebrowser[ext]', '$HOME/.filebrowser[ext]' or '/etc/filebrowser/.filebrowser[ext]')")
	vaddP(pf, "database", "d", "./filebrowser.db", "path to the database")
	vaddP(f, "address", "a", "127.0.0.1", "address to listen on")
	vaddP(f, "log", "l", "stdout", "log output")
	vaddP(f, "port", "p", 8080, "port to listen on")
	vaddP(f, "cert", "t", "", "tls certificate")
	vaddP(f, "key", "k", "", "tls key")
	vaddP(f, "scope", "s", ".", "scope to prepend to a user's scope when it is relative")

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
manage your user and all the configurations without accessing the
web interface.

If you've never run File Browser, you will need to create the database.
See 'filebrowser help config init' for more information.`,
	Run: serveAndListen,
}

func serveAndListen(cmd *cobra.Command, args []string) {
	initConfig()

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

	if _, err := os.Stat(v.GetString("database")); os.IsNotExist(err) {
		quickSetup(cmd)
	}

	db := getDB()
	defer db.Close()
	st := getStorage(db)

	port := v.GetInt("port")
	address := v.GetString("address")
	cert := v.GetString("cert")
	key := v.GetString("key")
	scope := v.GetString("scope")

	scope, err := filepath.Abs(scope)
	checkErr(err)
	settings, err := st.Settings.Get()
	checkErr(err)
	settings.Scope = scope
	err = st.Settings.Save(settings)
	checkErr(err)

	handler, err := fbhttp.NewHandler(st)
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
}

func quickSetup(cmd *cobra.Command) {
	db, err := storm.Open(v.GetString("database"))
	checkErr(err)
	defer db.Close()

	set := &settings.Settings{
		Key:        generateRandomBytes(64), // 256 bit
		BaseURL:    "",
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

	st := getStorage(db)

	err = st.Settings.Save(set)
	checkErr(err)

	err = st.Auth.Save(&auth.JSONAuth{})
	checkErr(err)

	password, err := users.HashPwd("admin")
	checkErr(err)

	user := &users.User{
		Username:     "admin",
		Password:     password,
		LockPassword: false,
	}

	set.Defaults.Apply(user)
	user.Perm.Admin = true

	err = st.Users.Save(user)
	checkErr(err)
}

// initConfig reads in config file and ENV variables if set.
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
		log.Println("No config file provided")
	} else {
		log.Println("Using config file:", v.ConfigFileUsed())
	}
}

package cmd

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/auth"
	

	fbhttp "github.com/filebrowser/filebrowser/http"
	"github.com/spf13/cobra"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	databasePath string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&databasePath, "database", "d", "./filebrowser.db", "path to the database")

	rootCmd.Flags().StringP("address", "a", "127.0.0.1", "address to listen on")
	rootCmd.Flags().StringP("log", "l", "stderr", "log output")
	rootCmd.Flags().IntP("port", "p", 0, "port to listen on")
	rootCmd.Flags().StringP("cert", "c", "", "tls certificate")
	rootCmd.Flags().StringP("key", "k", "", "tls key")
	rootCmd.Flags().StringP("scope", "s", "", "scope for users")
}

var rootCmd = &cobra.Command{
	Use:   "filebrowser",
	Short: "A stylish web-based file browser",
	Long: `File Browser CLI lets you create the database to use with File Browser,
manage your user and all the configurations without accessing the
web interface.

If you've never run File Browser, you will need to create the database.
See 'filebrowser help config init' for more information.

This command is used to start up the server. By default it starts
listening on loalhost on a random port. Use the flags to change it.`,
	Run: func(cmd *cobra.Command, args []string) {
		setupLogger(cmd)

		if _, err := os.Stat(databasePath); os.IsNotExist(err) {
			quickSetup(cmd)
		}

		var err error
		db := getDB()
		defer db.Close()
		fb := getFileBrowser(db)

		handler, err := fbhttp.NewHandler(fb)
		checkErr(err)
		startServer(cmd, handler)
	},
}

func setupLogger(cmd *cobra.Command) {
	switch l := mustGetString(cmd, "log"); l {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "":
		log.SetOutput(ioutil.Discard)
	default:
		log.SetOutput(&lumberjack.Logger{
			Filename:   l,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		})
	}
}

func quickSetup(cmd *cobra.Command) {
	scope := mustGetString(cmd, "scope")
	if scope == "" {
		panic(errors.New("scope flag must be set for quick setup"))
	}

	db, err := storm.Open(databasePath)
	checkErr(err)
	defer db.Close()
	fb := getFileBrowser(db)

	settings := fb.GetSettings()
	settings.BaseURL = ""
	settings.Signup = false
	settings.AuthMethod = auth.MethodJSONAuth
	settings.Defaults = lib.UserDefaults{
		Scope:  scope,
		Locale: "en",
		Perm: lib.Permissions{
			Admin:    false,
			Execute:  true,
			Create:   true,
			Rename:   true,
			Modify:   true,
			Delete:   true,
			Share:    true,
			Download: true,
		},
	}

	err = fb.SaveSettings(settings)
	checkErr(err)
	
	err = fb.SaveAuther(&auth.JSONAuth{})
	checkErr(err)
	
	password, err := lib.HashPwd("admin")
	checkErr(err)

	user := &lib.User{
		Username:     "admin",
		Password:     password,
		LockPassword: false,
	}

	fb.ApplyDefaults(user)
	user.Perm.Admin = true

	err = fb.SaveUser(user)
	checkErr(err)
}

func startServer(cmd *cobra.Command, handler http.Handler) {
	addr := mustGetString(cmd, "address")
	port, err := cmd.Flags().GetInt("port")
	checkErr(err)

	cert := mustGetString(cmd, "cert")
	key := mustGetString(cmd, "key")

	var listener net.Listener

	if cert != "" && key != "" {
		cer, err := tls.LoadX509KeyPair(cert, key)
		checkErr(err)
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		listener, err = tls.Listen("tcp", addr+":"+strconv.Itoa(port), config)
		checkErr(err)
	} else {
		listener, err = net.Listen("tcp", addr+":"+strconv.Itoa(port))
		checkErr(err)
	}

	log.Println("Listening on", listener.Addr().String())
	if err := http.Serve(listener, handler); err != nil {
		log.Fatal(err)
	}
}

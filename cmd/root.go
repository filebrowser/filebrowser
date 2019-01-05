package cmd

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"

	fbhttp "github.com/filebrowser/filebrowser/v2/http"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	databasePath string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&databasePath, "database", "d", "./filebrowser.db", "path to the database")

	rootCmd.Flags().StringP("address", "a", "", "address to listen on (default comes from database)")
	rootCmd.Flags().StringP("log", "l", "", "log output (default comes from database)")
	rootCmd.Flags().IntP("port", "p", 0, "port to listen on (default comes from database)")
	rootCmd.Flags().StringP("cert", "c", "", "tls certificate (default comes from database)")
	rootCmd.Flags().StringP("key", "k", "", "tls key (default comes from database)")
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

This command is used to start up the server. By default it starts listening
on localhost on a random port unless specified otherwise in the database or
via flags.

Use the available flags to override the database/default options. These flags
values won't be persisted to the database. To persist configuration to the database
use the command 'filebrowser config set'.`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(databasePath); os.IsNotExist(err) {
			quickSetup(cmd)
		}

		db := getDB()
		defer db.Close()
		st := getStorage(db)
		startServer(cmd, st)
	},
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

func serverVisitAndReplace(cmd *cobra.Command, s *settings.Settings) {
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		switch flag.Name {
		case "log":
			s.Log = mustGetString(cmd, flag.Name)
		case "address":
			s.Server.Address = mustGetString(cmd, flag.Name)
		case "port":
			s.Server.Port = mustGetInt(cmd, flag.Name)
		case "cert":
			s.Server.TLSCert = mustGetString(cmd, flag.Name)
		case "key":
			s.Server.TLSKey = mustGetString(cmd, flag.Name)
		}
	})
}

func quickSetup(cmd *cobra.Command) {
	scope := mustGetString(cmd, "scope")
	if scope == "" {
		panic(errors.New("scope flag must be set for quick setup"))
	}

	db, err := storm.Open(databasePath)
	checkErr(err)
	defer db.Close()

	set := &settings.Settings{
		Key:        generateRandomBytes(64), // 256 bit
		BaseURL:    "",
		Log:        "stderr",
		Signup:     false,
		AuthMethod: auth.MethodJSONAuth,
		Server: settings.Server{
			Port:    0,
			Address: "127.0.0.1",
			TLSCert: mustGetString(cmd, "cert"),
			TLSKey:  mustGetString(cmd, "key"),
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

	serverVisitAndReplace(cmd, set)
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

func startServer(cmd *cobra.Command, st *storage.Storage) {
	settings, err := st.Settings.Get()
	checkErr(err)

	serverVisitAndReplace(cmd, settings)
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

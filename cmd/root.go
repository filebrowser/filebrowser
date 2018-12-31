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

	"github.com/filebrowser/filebrowser/auth"
	"github.com/filebrowser/filebrowser/bolt"
	"github.com/filebrowser/filebrowser/types"

	fhttp "github.com/filebrowser/filebrowser/http"
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

		env := &fhttp.Env{
			Store: getStore(db),
		}

		env.Settings, err = env.Store.Config.GetSettings()
		checkErr(err)
		env.Auther, err = env.Store.Config.GetAuther(env.Settings.AuthMethod)
		checkErr(err)
		env.Runner, err = env.Store.Config.GetRunner()
		checkErr(err)

		startServer(cmd, env)
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

	settings := &types.Settings{
		Key:        generateRandomBytes(64),
		BaseURL:    "",
		Signup:     false,
		AuthMethod: auth.MethodJSONAuth,
		Defaults: types.UserDefaults{
			Scope:  scope,
			Locale: "en",
			Perm: types.Permissions{
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

	password, err := types.HashPwd("admin")
	checkErr(err)

	user := &types.User{
		Username:     "admin",
		Password:     password,
		LockPassword: false,
	}

	user.ApplyDefaults(settings.Defaults)
	user.Perm.Admin = true

	db, err := bolt.Open(databasePath)
	checkErr(err)
	defer db.Close()

	saveConfig(db, settings, &types.Runner{}, &auth.JSONAuth{})

	st := getStore(db)
	err = st.Users.Save(user)
	checkErr(err)
}

func startServer(cmd *cobra.Command, env *fhttp.Env) {
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
	if err := http.Serve(listener, fhttp.Handler(env)); err != nil {
		log.Fatal(err)
	}
}

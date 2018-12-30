package cmd

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/filebrowser/filebrowser/bolt"
	fhttp "github.com/filebrowser/filebrowser/http"
	"github.com/filebrowser/filebrowser/types"
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
	rootCmd.AddCommand(versionCmd)
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

		var err error
		db := getDB()
		defer db.Close()

		env := &fhttp.Env{
			Store: &types.Store{
				Users:  bolt.UsersStore{DB: db},
				Config: bolt.ConfigStore{DB: db, Users: bolt.UsersStore{DB: db}},
				Share:  bolt.ShareStore{DB: db},
			},
		}

		env.Settings, err = env.Store.Config.GetSettings()
		checkErr(err)
		env.Auther, err = env.Store.Config.GetAuther(env.Settings.AuthMethod)
		checkErr(err)
		env.Runner, err = env.Store.Config.GetRunner()
		checkErr(err)

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
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("File Browser Version " + types.Version)
	},
}

// Execute executes the commands.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

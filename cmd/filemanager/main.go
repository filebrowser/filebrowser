package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/hacdias/filemanager"
	"github.com/hacdias/fileutils"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	addr          string
	config        string
	database      string
	scope         string
	commands      string
	logfile       string
	port          int
	allowCommands bool
	allowEdit     bool
	allowNew      bool
)

func init() {
	flag.StringVarP(&config, "config", "c", "", "JSON configuration file")
	flag.IntVarP(&port, "port", "p", 0, "HTTP Port (default is random)")
	flag.StringVarP(&addr, "address", "a", "", "Address to listen to (default is all of them)")
	flag.StringVarP(&database, "database", "d", "./filemanager.db", "Database file")
	flag.StringVarP(&logfile, "log", "l", "stdout", "Errors logger; can use 'stdout', 'stderr' or file")
	flag.StringVarP(&scope, "scope", "s", ".", "Default scope option for new users")
	flag.StringVar(&commands, "commands", "git svn hg", "Default commands option for new users")
	flag.BoolVar(&allowCommands, "allow-commands", true, "Default allow commands option for new users")
	flag.BoolVar(&allowEdit, "allow-edit", true, "Default allow edit option for new users")
	flag.BoolVar(&allowNew, "allow-new", true, "Default allow new option for new users")
}

func setupViper() {
	viper.SetDefault("Address", "")
	viper.SetDefault("Port", "0")
	viper.SetDefault("Database", "./filemanager.db")
	viper.SetDefault("Scope", ".")
	viper.SetDefault("Logger", "stdout")
	viper.SetDefault("Commands", []string{"git", "svn", "hg"})
	viper.SetDefault("AllowCommmands", true)
	viper.SetDefault("AllowEdit", true)
	viper.SetDefault("AllowNew", true)

	viper.BindPFlag("Port", flag.Lookup("port"))
	viper.BindPFlag("Address", flag.Lookup("address"))
	viper.BindPFlag("Database", flag.Lookup("database"))
	viper.BindPFlag("Scope", flag.Lookup("scope"))
	viper.BindPFlag("Logger", flag.Lookup("log"))
	viper.BindPFlag("Commands", flag.Lookup("commands"))
	viper.BindPFlag("AllowCommands", flag.Lookup("allow-commands"))
	viper.BindPFlag("AllowEdit", flag.Lookup("allow-edit"))
	viper.BindPFlag("AlowNew", flag.Lookup("allow-new"))

	viper.SetConfigName("filemanager")
	viper.AddConfigPath(".")
}

func main() {
	setupViper()
	flag.Parse()

	// Add a configuration file if set.
	if config != "" {
		viper.SetConfigName(strings.TrimSuffix(config, filepath.Ext(config)))
	}

	// Read configuration from a file if exists.
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigParseError); ok {
			panic(err)
		}
	}

	// Set up process log before anything bad happens.
	switch viper.GetString("Logger") {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "":
		log.SetOutput(ioutil.Discard)
	default:
		log.SetOutput(&lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		})
	}

	// Create a File Manager instance.
	fm, err := filemanager.New(viper.GetString("Database"), filemanager.User{
		AllowCommands: viper.GetBool("AllowCommands"),
		AllowEdit:     viper.GetBool("AllowEdit"),
		AllowNew:      viper.GetBool("AllowNew"),
		Commands:      viper.GetStringSlice("Commands"),
		Rules:         []*filemanager.Rule{},
		CSS:           "",
		FileSystem:    fileutils.Dir(viper.GetString("Scope")),
	})

	if err != nil {
		log.Fatal(err)
	}

	// Builds the address and a listener.
	laddr := viper.GetString("Address") + ":" + viper.GetString("Port")
	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}

	// Tell the user the port in which is listening.
	fmt.Println("Listening on", listener.Addr().String())

	// Starts the server.
	if err := http.Serve(listener, fm); err != nil {
		log.Fatal(err)
	}
}

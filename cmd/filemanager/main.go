package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/hacdias/filemanager"

	"golang.org/x/net/webdav"
)

// confFile contains the configuration file for this File Manager instance.
// If the user chooses to use a configuration file, the flags will be ignored.
type confFile struct {
	Database      string   `json:"database"`
	Scope         string   `json:"scope"`
	Commands      []string `json:"commands"`
	Port          int      `json:"port"`
	AllowCommands bool     `json:"allowCommands"`
	AllowEdit     bool     `json:"allowEdit"`
	AllowNew      bool     `json:"allowNew"`
}

var (
	config        string
	database      string
	scope         string
	commands      string
	port          string
	allowCommands bool
	allowEdit     bool
	allowNew      bool
)

func init() {
	flag.StringVar(&config, "config", "", "JSON configuration file")
	flag.StringVar(&port, "port", "80", "HTTP Port")
	flag.StringVar(&database, "database", "./filemanager.db", "Database path")
	flag.StringVar(&scope, "scope", ".", "Defualt scope for new users")
	flag.StringVar(&commands, "commands", "git svn hg", "Space separated commands available for new users")
	flag.BoolVar(&allowCommands, "allow-commands", true, "Default allow commands option")
	flag.BoolVar(&allowEdit, "allow-edit", true, "Default allow edit option")
	flag.BoolVar(&allowNew, "allow-new", true, "Default allow new option")
}

func main() {
	flag.Parse()

	if config != "" {
		loadConfig()
	}

	fm, err := filemanager.New(database, filemanager.User{
		Username:      "admin",
		Password:      "admin",
		AllowCommands: allowCommands,
		AllowEdit:     allowEdit,
		AllowNew:      allowNew,
		Commands:      strings.Split(strings.TrimSpace(commands), " "),
		Rules:         []*filemanager.Rule{},
		CSS:           "",
		FileSystem:    webdav.Dir(scope),
	})

	if err != nil {
		panic(err)
	}

	fm.SetBaseURL("/")
	fm.SetPrefixURL("/")

	fmt.Println("Starting filemanager on *:" + port)
	if err := http.ListenAndServe(":"+port, fm); err != nil {
		panic(err)
	}
}

func loadConfig() {
	file, err := ioutil.ReadFile(config)
	if err != nil {
		panic(err)
	}

	var conf *confFile
	err = json.Unmarshal(file, &conf)
	if err != nil {
		panic(err)
	}

	database = conf.Database
	scope = conf.Scope
	commands = strings.Join(conf.Commands, " ")
	port = strconv.Itoa(conf.Port)
	allowNew = conf.AllowNew
	allowEdit = conf.AllowEdit
	allowCommands = conf.AllowCommands
}

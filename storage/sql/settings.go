package sql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

type settingsBackend struct {
	db *sql.DB
}

func InitSettingsTable(db *sql.DB) error {
	sql := "create table if not exists settings(key string primary key, value string)"
	_, err := db.Exec(sql)
	return err
}

func userDefaultsFromString(s string) settings.UserDefaults {
	if s == "" {
		return settings.UserDefaults{}
	}
	userDefaults := settings.UserDefaults{}
	err := json.Unmarshal([]byte(s), &userDefaults)
	if err != nil {
		fmt.Println("ERROR: Fail to parse settings.UserDefaults")
	}
	return userDefaults
}

func userDefaultsToString(d settings.UserDefaults) string {
	data, err := json.Marshal(d)
	if err != nil {
		return ""
	}
	return string(data)
}

func brandingFromString(s string) settings.Branding {
	if s == "" {
		return settings.Branding{}
	}
	branding := settings.Branding{}
	err := json.Unmarshal([]byte(s), &branding)
	if err != nil {
		fmt.Println("ERROR: Fail to parse settings.Branding")
	}
	return branding
}

func brandingToString(s settings.Branding) string {
	data, err := json.Marshal(s)
	if err != nil {
		fmt.Println("ERROR: Fail to jsonify settings.Branding")
		return ""
	}
	return string(data)
}

func commandsToString(c map[string][]string) string {
	data, err := json.Marshal(c)
	if err != nil {
		fmt.Println("ERROR: Fail to jsonify commands")
		return ""
	}
	return string(data)
}

func commandsFromString(s string) map[string][]string {
	if s == "" {
		return map[string][]string{}
	}
	c := map[string][]string{}
	err := json.Unmarshal([]byte(s), &c)
	if err != nil {
		fmt.Println("ERROR: Fail to parse commands")
	}
	return c
}

func stringsFromString(s string) []string {
	if s == "" {
		return []string{}
	}
	c := []string{}
	err := json.Unmarshal([]byte(s), &c)
	if err != nil {
		fmt.Println("ERROR: Fail to parse []string")
	}
	return c
}

func stringsToString(c []string) string {
	data, err := json.Marshal(c)
	if err != nil {
		fmt.Println("ERROR: Fail to jsonify strings")
		return ""
	}
	return string(data)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func boolFromInt(i int) bool {
	if i == 0 {
		return false
	}
	return true
}

func boolFromString(s string) bool {
	if s == "0" || s == "" || s == "f" || s == "F" {
		return false
	}
	return true
}

func boolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (s settingsBackend) Get() (*settings.Settings, error) {
	sql := "select key, value from settings"
	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, nil
	}
	key := ""
	value := ""
	settings1 := cloneSettings(defaultSettings)
	for rows.Next() {
		err = rows.Scan(&key, &value)
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("ERROR: Fail to query settings.Settings")
		}
		if key == "Key" {
			settings1.Key = []byte(value)
		} else if key == "Signup" {
			settings1.Signup = boolFromString(value)
		} else if key == "CreateUserDir" {
			settings1.CreateUserDir = boolFromString(value)
		} else if key == "Defaults" {
			settings1.Defaults = userDefaultsFromString(value)
		} else if key == "AuthMethod" {
			settings1.AuthMethod = settings.AuthMethod(value)
		} else if key == "Branding" {
			settings1.Branding = brandingFromString(value)
		} else if key == "Commands" {
			settings1.Commands = commandsFromString(value)
		} else if key == "Shell" {
			settings1.Shell = stringsFromString(value)
		} else if key == "Rules" {
			settings1.Rules = rulesFromString(value)
		}
	}
	return &settings1, nil
}

func (s settingsBackend) Save(ss *settings.Settings) error {
	columns := []string{"Key", "Signup", "CreateUserDir", "UserHomeBasePath", "Defaults", "AuthMethod", "Branding", "Commands", "Shell", "Rules"}
	values := []string{
		"'" + string(ss.Key) + "'",
		boolToString(ss.Signup),
		boolToString(ss.CreateUserDir),
		"'" + string(ss.UserHomeBasePath) + "'",
		userDefaultsToString(ss.Defaults),
		string(ss.AuthMethod),
		brandingToString(ss.Branding),
		commandsToString(ss.Commands),
		stringsToString(ss.Shell),
		RulesToString(ss.Rules)}
	sql := fmt.Sprintf("INSERT INTO settings (%s) VALUES(%s)", strings.Join(columns, ","), strings.Join(values, ","))
	_, err := s.db.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

var defaultServer = settings.Server{
	Port:                  "8080",
	Log:                   "stdout",
	EnableThumbnails:      false,
	ResizePreview:         false,
	EnableExec:            false,
	TypeDetectionByHeader: false,
}

var defaultSettings = settings.Settings{
	Key:              []byte(""),
	Signup:           false,
	CreateUserDir:    false,
	UserHomeBasePath: "/users",
	Defaults: settings.UserDefaults{
		Scope:       ".",
		Locale:      "en",
		ViewMode:    "mosaic",
		SingleClick: false,
		Sorting: files.Sorting{
			By:  "",
			Asc: false,
		},
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
		Commands:     []string{},
		HideDotfiles: false,
		DateFormat:   false,
	},
	AuthMethod: "json",
	Branding: settings.Branding{
		Name:            "",
		DisableExternal: false,
		Files:           "",
		Theme:           "",
		Color:           "",
	},
	Commands: map[string][]string{},
	Shell:    []string{},
	Rules:    []rules.Rule{},
}

func cloneServer(server settings.Server) settings.Server {
	data, err := json.Marshal(server)
	if err != nil {
		return settings.Server{}
	}
	s := settings.Server{}
	json.Unmarshal(data, &s)
	return s
}

func cloneSettings(s settings.Settings) settings.Settings {
	data, err := json.Marshal(s)
	if err != nil {
		return settings.Settings{}
	}
	s1 := settings.Settings{}
	json.Unmarshal(data, &s1)
	return s1
}

func (s settingsBackend) GetServer() (*settings.Server, error) {
	sql := "select key, value from settings"
	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, nil
	}
	server := cloneServer(defaultServer)
	key := ""
	value := ""

	for rows.Next() {
		err = rows.Scan(&key, &value)
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("ERROR: Fail to query settings.Settings")
		}
		if key == "Root" {
			server.Root = value
		} else if key == "BaseURL" {
			server.BaseURL = value
		} else if key == "Socket" {
			server.Socket = value
		} else if key == "TLSKey" {
			server.TLSKey = value
		} else if key == "TLSCert" {
			server.TLSCert = value
		} else if key == "Port" {
			server.Port = value
		} else if key == "Address" {
			server.Address = value
		} else if key == "Log" {
			server.Log = value
		} else if key == "EnableThumbnails" {
			server.EnableThumbnails = boolFromString(value)
		} else if key == "ResizePreview" {
			server.ResizePreview = boolFromString(value)
		} else if key == "EnableExec" {
			server.EnableExec = boolFromString(value)
		} else if key == "TypeDetectionByHeader" {
			server.TypeDetectionByHeader = boolFromString(value)
		} else if key == "AuthHook" {
			server.AuthHook = value
		}
	}
	return &server, nil
}

func (s settingsBackend) SaveServer(ss *settings.Server) error {
	columns := []string{"Root", "BaseURL", "Socket", "TLSKey", "TLSCert", "Port", "Address", "Log", "EnableThumbnails", "ResizePreview", "EnableExec", "TypeDetectionByHeader", "AuthHook"}
	values := []string{
		"'" + ss.Root + "'",
		"'" + ss.BaseURL + "'",
		"'" + ss.Socket + "'",
		"'" + ss.TLSKey + "'",
		"'" + ss.TLSCert + "'",
		"'" + ss.Port + "'",
		"'" + ss.Address + "'",
		"'" + ss.Log + "'",
		boolToString(ss.EnableThumbnails),
		boolToString(ss.ResizePreview),
		boolToString(ss.EnableExec),
		boolToString(ss.TypeDetectionByHeader),
		"'" + ss.AuthHook + "'"}
	sql := fmt.Sprintf("INSERT INTO settings (%s) VALUES(%s)", strings.Join(columns, ","), strings.Join(values, ","))
	_, err := s.db.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func SetSetting(db *sql.DB, key string, value string) error {
	sql := "select count(key) from settings where key = '" + key + "'"
	count := 0
	err := db.QueryRow(sql).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return addSetting(db, key, value)
	}
	return updateSetting(db, key, value)
}

func GetSetting(db *sql.DB, key string) string {
	sql := "select value from settings where key = '" + key + "';"
	value := ""
	err := db.QueryRow(sql).Scan(&value)
	if err != nil {
		return value
	}
	return value
}

func addSetting(db *sql.DB, key string, value string) error {
	sql := "insert into settings(key, value) values('" + key + "', '" + value + "')"
	_, err := db.Exec(sql)
	return err
}

func updateSetting(db *sql.DB, key string, value string) error {
	sql := "update settings set value = '" + value + "' where key = '" + key + "'"
	_, err := db.Exec(sql)
	return err
}

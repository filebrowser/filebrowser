package sql

import (
	"database/sql"
	"encoding/json"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

func init() {
}

type settingsBackend struct {
	db *sql.DB
}

func InitSettingsTable(db *sql.DB) error {
	sql := "create table if not exists settings(key string primary key, value string)"
	_, err := db.Exec(sql)
	checkError(err, "Fail to create table settings")
	return err
}

func userDefaultsFromString(s string) settings.UserDefaults {
	if s == "" {
		return settings.UserDefaults{}
	}
	userDefaults := settings.UserDefaults{}
	err := json.Unmarshal([]byte(s), &userDefaults)
	checkError(err, "Fail to parse settings.UserDefaults")
	return userDefaults
}

func userDefaultsToString(d settings.UserDefaults) string {
	data, err := json.Marshal(d)
	if checkError(err, "Fail to stringify settings.UserDefaults") {
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
	checkError(err, "Fail to parse settings.Branding")
	return branding
}

func brandingToString(s settings.Branding) string {
	data, err := json.Marshal(s)
	if checkError(err, "Fail to jsonify settings.Branding") {
		return ""
	}
	return string(data)
}

func commandsToString(c map[string][]string) string {
	data, err := json.Marshal(c)
	if checkError(err, "Fail to jsonify commands") {
		return ""
	}
	return string(data)
}

func commandsFromString(s string) map[string][]string {
	c := make(map[string][]string)
	if s == "" {
		return c
	}
	err := json.Unmarshal([]byte(s), &c)
	checkError(err, "Fail to parse commands")
	return c
}

func stringsFromString(s string) []string {
	c := make([]string, 0)
	if s == "" {
		return c
	}
	err := json.Unmarshal([]byte(s), &c)
	checkError(err, "Fail to parse []string")
	return c
}

func stringsToString(c []string) string {
	data, err := json.Marshal(c)
	if checkError(err, "Fail to jsonify strings") {
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
	logBacktrace()
	sql := "select key, value from settings"
	rows, err := s.db.Query(sql)
	if checkError(err, "Fail to Query settings.Settings") {
		return nil, err
	}
	key := ""
	value := ""
	settings1 := cloneSettings(defaultSettings)
	for rows.Next() {
		err = rows.Scan(&key, &value)
		checkError(err, "Fail to query settings.Settings")
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
	if len(settings1.Key) == 0 {
		return nil, nil
	}
	return &settings1, nil
}

func (s settingsBackend) Save(ss *settings.Settings) error {
	logBacktrace()
	fields := []string{"Key", "Signup", "CreateUserDir", "UserHomeBasePath", "Defaults", "AuthMethod", "Branding", "Commands", "Shell", "Rules"}
	values := []string{
		string(ss.Key),
		boolToString(ss.Signup),
		boolToString(ss.CreateUserDir),
		string(ss.UserHomeBasePath),
		userDefaultsToString(ss.Defaults),
		string(ss.AuthMethod),
		brandingToString(ss.Branding),
		commandsToString(ss.Commands),
		stringsToString(ss.Shell),
		RulesToString(ss.Rules),
	}
	tx, err := s.db.Begin()
	if checkError(err, "Fail to begin db transaction") {
		return err
	}
	for i, field := range fields {
		stmt, err := s.db.Prepare("INSERT INTO settings (key, value) VALUES(?,?)")
		defer stmt.Close()
		if checkError(err, "Fail to prepare statement") {
			tx.Rollback()
			break
		}
		_, err = stmt.Exec(field, values[i])
		if checkError(err, "Fail to insert field "+field+" of settings") {
			tx.Rollback()
			break
		}
	}
	err = tx.Commit()
	if checkError(err, "Fail to commit") {
		tx.Rollback()
		return err
	}
	return err
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
		Commands:     make([]string, 0),
		HideDotfiles: false,
		DateFormat:   false,
	},
	AuthMethod: auth.MethodJSONAuth,
	Branding: settings.Branding{
		Name:            "",
		DisableExternal: false,
		Files:           "",
		Theme:           "",
		Color:           "",
	},
	Commands: make(map[string][]string),
	Shell:    make([]string, 0),
	Rules:    make([]rules.Rule, 0),
}

func cloneServer(server settings.Server) settings.Server {
	data, err := json.Marshal(server)
	s := settings.Server{}
	if checkError(err, "Fail to clone settings.Server") {
		return s
	}
	err = json.Unmarshal(data, &s)
	checkError(err, "Fail to decode for settings.Server")
	return s
}

func cloneSettings(s settings.Settings) settings.Settings {
	data, err := json.Marshal(s)
	s1 := settings.Settings{}
	if checkError(err, "Fail to clone settings.Settings") {
		return s1
	}
	json.Unmarshal(data, &s1)
	return s1
}

func (s settingsBackend) GetServer() (*settings.Server, error) {
	logBacktrace()
	sql := "select key, value from settings"
	rows, err := s.db.Query(sql)
	if checkError(err, "Fail to Query for GetServer") {
		return nil, err
	}
	server := cloneServer(defaultServer)
	key := ""
	value := ""

	for rows.Next() {
		err = rows.Scan(&key, &value)
		if checkError(err, "Fail to query settings.Settings") {
			continue
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
	logBacktrace()
	fields := []string{"Root", "BaseURL", "Socket", "TLSKey", "TLSCert", "Port", "Address", "Log", "EnableThumbnails", "ResizePreview", "EnableExec", "TypeDetectionByHeader", "AuthHook"}
	values := []string{
		ss.Root,
		ss.BaseURL,
		ss.Socket,
		ss.TLSKey,
		ss.TLSCert,
		ss.Port,
		ss.Address,
		ss.Log,
		boolToString(ss.EnableThumbnails),
		boolToString(ss.ResizePreview),
		boolToString(ss.EnableExec),
		boolToString(ss.TypeDetectionByHeader),
		ss.AuthHook}
	tx, err := s.db.Begin()
	if checkError(err, "Fail to begin db transaction") {
		return err
	}
	for i, field := range fields {
		stmt, err := s.db.Prepare("INSERT INTO settings (key, value) VALUES(?,?)")
		defer stmt.Close()
		if checkError(err, "Fail to prepare statement") {
			tx.Rollback()
			break
		}
		_, err = stmt.Exec(field, values[i])
		if checkError(err, "Fail to insert field "+field+" of settings") {
			tx.Rollback()
			break
		}
	}
	err = tx.Commit()
	if checkError(err, "Fail to commit") {
		tx.Rollback()
		return err
	}
	return err
}

func SetSetting(db *sql.DB, key string, value string) error {
	sql := "select count(key) from settings where key = '" + key + "'"
	count := 0
	err := db.QueryRow(sql).Scan(&count)
	if checkError(err, "Fail to QueryRow for key="+key) {
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
	if checkError(err, "Fail to QueryRow for key "+key) {
		return value
	}
	return value
}

func addSetting(db *sql.DB, key string, value string) error {
	sql := "insert into settings(key, value) values('" + key + "', '" + value + "')"
	_, err := db.Exec(sql)
	checkError(err, "Fail to addSetting")
	return err
}

func updateSetting(db *sql.DB, key string, value string) error {
	sql := "update settings set value = '" + value + "' where key = '" + key + "'"
	_, err := db.Exec(sql)
	checkError(err, "Fail to updateSetting")
	return err
}

func HadSetting(db *sql.DB) bool {
	key := GetSetting(db, "Key")
	if key == "" {
		return false
	}
	return true
}

package sql

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

func init() {
}

type settingsBackend struct {
	db     *sql.DB
	dbType string
}

func InitSettingsTable(db *sql.DB, dbType string) error {
	sql := fmt.Sprintf("create table if not exists %s (%s varchar(128) primary key, value text);", quoteName(dbType, SettingsTable), quoteName(dbType, "key"))
	_, err := db.Exec(sql)
	checkError(err, "Fail to create table settings")
	return err
}

func bytesToString(data []byte) string {
	return base64.RawStdEncoding.EncodeToString(data)
}

func bytesFromString(s string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(s)
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
	sql := fmt.Sprintf("select %s, value from %s;", quoteName(s.dbType, "key"), quoteName(s.dbType, SettingsTable))
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
			val, err := bytesFromString(value)
			if !checkError(err, "Fail to parse []byte from string") {
				settings1.Key = val
			}
		} else if key == "Signup" {
			settings1.Signup = boolFromString(value)
		} else if key == "CreateUserDir" {
			settings1.CreateUserDir = boolFromString(value)
		} else if key == "UserHomeBasePath" {
			settings1.UserHomeBasePath = value
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
	fields := []string{"Key", "Signup", "CreateUserDir", "UserHomeBasePath", "Defaults", "AuthMethod", "Branding", "Commands", "Shell", "Rules"}
	values := []string{
		bytesToString(ss.Key),
		boolToString(ss.Signup),
		boolToString(ss.CreateUserDir),
		ss.UserHomeBasePath,
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
	table := quoteName(s.dbType, SettingsTable)
	k := quoteName(s.dbType, "key")
	p1 := placeHolder(s.dbType, 1)
	p2 := placeHolder(s.dbType, 2)
	for i, field := range fields {
		exists := ContainKey(s.db, s.dbType, field)
		sql := fmt.Sprintf("INSERT INTO %s (value, %s) VALUES(%s,%s);", table, k, p1, p2)
		if exists {
			sql = fmt.Sprintf("UPDATE %s set value = %s where %s = %s;", table, p1, k, p2)
		}
		stmt, err := s.db.Prepare(sql)
		defer stmt.Close()
		if checkError(err, "Fail to prepare statement") {
			tx.Rollback()
			break
		}
		_, err = stmt.Exec(values[i], field)
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

func cloneSettings(s settings.Settings) settings.Settings {
	data, err := json.Marshal(s)
	s1 := settings.Settings{}
	if checkError(err, "Fail to clone settings.Settings") {
		return s1
	}
	json.Unmarshal(data, &s1)
	return s1
}

func SetSetting(db *sql.DB, dbType string, key string, value string) error {
	t := quoteName(dbType, SettingsTable)
	k := quoteName(dbType, "key")
	sql := fmt.Sprintf("select count(%s) from %s where %s = '%s';", k, t, k, key)
	count := 0
	err := db.QueryRow(sql).Scan(&count)
	if checkError(err, "Fail to QueryRow for key="+key) {
		return err
	}
	if count == 0 {
		return addSetting(db, dbType, key, value)
	}
	return updateSetting(db, dbType, key, value)
}

func GetSetting(db *sql.DB, dbType string, key string) string {
	sql := fmt.Sprintf("select value from %s where %s = '%s';", quoteName(dbType, SettingsTable), quoteName(dbType, "key"), key)
	value := ""
	err := db.QueryRow(sql).Scan(&value)
	if checkError(err, "") {
		return value
	}
	return value
}

func addSetting(db *sql.DB, dbType string, key string, value string) error {
	table := quoteName(dbType, SettingsTable)
	k := quoteName(dbType, "key")
	p1 := placeHolder(dbType, 1)
	p2 := placeHolder(dbType, 2)
	sql := fmt.Sprintf("insert into %s (%s, value) values(%s, %s);", table, k, p1, p2)
	stmt, err := db.Prepare(sql)
	if checkError(err, "Fail to prepare sql") {
		return err
	}
	_, err = stmt.Exec(key, value)
	checkError(err, "Fail to add settings")
	return err
}

func updateSetting(db *sql.DB, dbType string, key string, value string) error {
	sql := fmt.Sprintf(
		"update %s set value = %s where %s = %s;",
		quoteName(dbType, SettingsTable),
		placeHolder(dbType, 1),
		quoteName(dbType, "key"),
		placeHolder(dbType, 2),
	)
	stmt, err := db.Prepare(sql)
	if checkError(err, "Fail to prepare sql") {
		return err
	}
	_, err = stmt.Exec(key, value)
	checkError(err, "Fail to updateSetting")
	return err
}

func HadSetting(db *sql.DB) bool {
	dbType, err := GetDBType(db)
	if checkError(err, "Fail to get db type") {
		return false
	}
	key := GetSetting(db, dbType, "Key")
	if key == "" {
		return false
	}
	return true
}

func ContainKey(db *sql.DB, dbType string, key string) bool {
	sql := fmt.Sprintf("select value from %s where %s = '%s';", quoteName(dbType, SettingsTable), quoteName(dbType, "key"), key)
	value := ""
	err := db.QueryRow(sql).Scan(&value)
	if checkError(err, "") {
		return false
	}
	return true
}

func HadSettingOfKey(db *sql.DB, dbType string, key string) bool {
	return GetSetting(db, dbType, "Key") == key
}

func newSettingsBackend(db *sql.DB, dbType string) settingsBackend {
	InitSettingsTable(db, dbType)
	return settingsBackend{db: db, dbType: dbType}
}

package sql

import "os"

var SettingsTable = "fb_settings"
var UsersTable = "fb_users"
var SharesTable = "fb_shares"

func getEnv(key string, defaultValue string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		return defaultValue
	}
	return val
}

func init() {
	SettingsTable = getEnv("FILEBROWSER_SETTINGS_TABLE", SettingsTable)
	UsersTable = getEnv("FILEBROWSER_USERS_TABLE", UsersTable)
	SharesTable = getEnv("FILEBROWSER_SHARES_TABLE", SharesTable)
}

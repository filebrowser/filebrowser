package users

import (
	"embed"
	"strings"
)

//go:embed assets
var assets embed.FS
var commonPasswords map[string]struct{}

//nolint:gochecknoinits
func init() {
	// Password list sourced from:
	// https://github.com/danielmiessler/SecLists/blob/master/Passwords/Common-Credentials/100k-most-used-passwords-NCSC.txt
	data, err := assets.ReadFile("assets/common-passwords.txt")
	if err != nil {
		panic(err)
	}

	passwords := strings.Split(strings.TrimSpace(string(data)), "\n")
	commonPasswords = make(map[string]struct{}, len(passwords))
	for _, password := range passwords {
		commonPasswords[strings.TrimSpace(password)] = struct{}{}
	}
}

package config

import (
	"strings"

	"golang.org/x/net/webdav"
)

// User contains the configuration for each user
type User struct {
	Scope         string            `json:"-"` // Path the user have access
	FileSystem    webdav.FileSystem `json:"-"` // The virtual file system the user have access
	Handler       *webdav.Handler   `json:"-"` // The WebDav HTTP Handler
	StyleSheet    string            `json:"-"` // Costum stylesheet
	AllowNew      bool              // Can create files and folders
	AllowEdit     bool              // Can edit/rename files
	AllowCommands bool              // Can execute commands
	Commands      []string          // Available Commands
	Rules         []*Rule           `json:"-"` // Access rules
}

// Allowed checks if the user has permission to access a directory/file
func (u User) Allowed(url string) bool {
	var rule *Rule
	i := len(u.Rules) - 1

	for i >= 0 {
		rule = u.Rules[i]

		if rule.Regex {
			if rule.Regexp.MatchString(url) {
				return rule.Allow
			}
		} else if strings.HasPrefix(url, rule.Path) {
			return rule.Allow
		}

		i--
	}

	return true
}

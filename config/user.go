package config

import (
	"net/http"
	"strings"
)

// User contains the configuration for each user
type User struct {
	PathScope     string          `json:"-"` // Path the user have access
	Root          http.FileSystem `json:"-"` // The virtual file system the user have access
	StyleSheet    string          `json:"-"` // Costum stylesheet
	FrontMatter   string          `json:"-"` // Default frontmatter to save files in
	AllowNew      bool            // Can create files and folders
	AllowEdit     bool            // Can edit/rename files
	AllowCommands bool            // Can execute commands
	Commands      []string        // Available Commands
	Rules         []*Rule         `json:"-"` // Access rules
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

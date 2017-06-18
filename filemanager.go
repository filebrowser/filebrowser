package filemanager

import (
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/webdav"
)

// CommandFunc ...
type CommandFunc func(r *http.Request, c *Config, u *User) error

// Config is a configuration for browsing in a particular path.
type Config struct {
	*User
	PrefixURL   string
	BaseURL     string
	WebDavURL   string
	HugoEnabled bool // Enables the Hugo plugin for File Manager
	Users       map[string]*User
	BeforeSave  CommandFunc
	AfterSave   CommandFunc
}

// AbsoluteURL ...
func (c Config) AbsoluteURL() string {
	return c.PrefixURL + c.BaseURL
}

// AbsoluteWebdavURL ...
func (c Config) AbsoluteWebdavURL() string {
	return c.PrefixURL + c.WebDavURL
}

// Rule is a dissalow/allow rule
type Rule struct {
	Regex  bool
	Allow  bool
	Path   string
	Regexp *regexp.Regexp
}

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

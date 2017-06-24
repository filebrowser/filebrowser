package filemanager

import (
	"net/http"
	"regexp"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"golang.org/x/net/webdav"
)

// FileManager is a file manager instance.
type FileManager struct {
	*User  `json:"-"`
	Assets *Assets `json:"-"`

	// PrefixURL is a part of the URL that is trimmed from the http.Request.URL before
	// it arrives to our handlers. It may be useful when using FileManager as a middleware
	// such as in caddy-filemanager plugin.
	PrefixURL string

	// BaseURL is the path where the GUI will be accessible.
	BaseURL string

	// WebDavURL is the path where the WebDAV will be accessible. It can be set to "/"
	// in order to override the GUI and only use the WebDAV.
	WebDavURL string

	// Users is a map with the different configurations for each user.
	Users map[string]*User `json:"-"`

	// TODO: event-based?
	BeforeSave CommandFunc `json:"-"`
	AfterSave  CommandFunc `json:"-"`
}

// User contains the configuration for each user.
type User struct {
	Scope         string            `json:"-"` // Path the user have access
	FileSystem    webdav.FileSystem `json:"-"` // The virtual file system the user have access
	Handler       *webdav.Handler   `json:"-"` // The WebDav HTTP Handler
	Rules         []*Rule           `json:"-"` // Access rules
	StyleSheet    string            `json:"-"` // Costum stylesheet
	AllowNew      bool              // Can create files and folders
	AllowEdit     bool              // Can edit/rename files
	AllowCommands bool              // Can execute commands
	Commands      []string          // Available Commands
}

// Assets are the static and front-end assets, such as JS, CSS and HTML templates.
type Assets struct {
	requiredJS rice.Box // JS that is always required to have in order to be usable.
	Templates  rice.Box
	CSS        rice.Box
	JS         rice.Box
}

// Rule is a dissalow/allow rule.
type Rule struct {
	Regex  bool
	Allow  bool
	Path   string
	Regexp *regexp.Regexp
}

// CommandFunc ...
type CommandFunc func(r *http.Request, c *FileManager, u *User) error

// AbsoluteURL ...
func (m FileManager) AbsoluteURL() string {
	return m.PrefixURL + m.BaseURL
}

// AbsoluteWebdavURL ...
func (m FileManager) AbsoluteWebdavURL() string {
	return m.PrefixURL + m.WebDavURL
}

// Allowed checks if the user has permission to access a directory/file.
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

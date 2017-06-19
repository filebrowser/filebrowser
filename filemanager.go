package filemanager

import (
	"net/http"
	"regexp"
	"strings"

	rice "github.com/GeertJohan/go.rice"

	"golang.org/x/net/webdav"
)

// FileManager is a configuration for browsing in a particular path.
type FileManager struct {
	*User
	PrefixURL  string // A part of the URL that is stripped from the http.Request
	BaseURL    string // The base URL of FileManager interface
	WebDavURL  string // The URL of WebDAV
	Users      map[string]*User
	BeforeSave Command
	AfterSave  Command
	Assets     struct {
		Templates *rice.Box
		Static    *rice.Box
	}
}

// New creates a new FileManager object with the default settings
// for a certain scope.
func New(scope string) *FileManager {
	fm := &FileManager{
		User: &User{
			Scope:         scope,
			FileSystem:    webdav.Dir(scope),
			AllowCommands: true,
			AllowEdit:     true,
			AllowNew:      true,
			Commands:      []string{"git", "svn", "hg"},
			Rules: []*Rule{{
				Regex:  true,
				Allow:  false,
				Regexp: regexp.MustCompile("\\/\\..+"),
			}},
		},
		Users:      map[string]*User{},
		BaseURL:    "",
		PrefixURL:  "",
		WebDavURL:  "/webdav",
		BeforeSave: func(r *http.Request, c *FileManager, u *User) error { return nil },
		AfterSave:  func(r *http.Request, c *FileManager, u *User) error { return nil },
	}

	fm.Handler = &webdav.Handler{
		Prefix:     fm.WebDavURL,
		FileSystem: fm.FileSystem,
		LockSystem: webdav.NewMemLS(),
	}

	return fm
}

// AbsoluteURL ...
func (c FileManager) AbsoluteURL() string {
	return c.PrefixURL + c.BaseURL
}

// AbsoluteWebdavURL ...
func (c FileManager) AbsoluteWebdavURL() string {
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
	Rules         []*Rule           `json:"-"` // Access rules
	AllowNew      bool              // Can create files and folders
	AllowEdit     bool              // Can edit/rename files
	AllowCommands bool              // Can execute commands
	Commands      []string          // Available Commands
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

// Command is a user-defined command that is executed in some moments.
type Command func(r *http.Request, c *FileManager, u *User) error

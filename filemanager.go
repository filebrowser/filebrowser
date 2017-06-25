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
	*User
	Assets *assets

	// PrefixURL is a part of the URL that is trimmed from the http.Request.URL before
	// it arrives to our handlers. It may be useful when using FileManager as a middleware
	// such as in caddy-filemanager plugin. It musn't end with a trailing slash.
	PrefixURL string

	// BaseURL is the path where the GUI will be accessible. It musn't end with
	// a trailing slash and mustn't contain PrefixURL, if set.
	BaseURL string

	// WebDavURL is the path where the WebDAV will be accessible. It can be set to ""
	// in order to override the GUI and only use the WebDAV. It musn't end with
	// a trailing slash.
	WebDavURL string

	scopes map[string]*scope

	// Users is a map with the different configurations for each user.
	Users map[string]*User

	// TODO: event-based?
	BeforeSave CommandFunc
	AfterSave  CommandFunc
}

type scope struct {
	path       string
	fileSystem webdav.FileSystem
	handler    *webdav.Handler
}

// User contains the configuration for each user.
type User struct {
	// scope is the physical path the user has access to.
	scope *scope

	// fileSystem is the virtual file system the user has access.
	fileSystem webdav.FileSystem

	// handler handles incoming requests to the WebDAV backend.
	handler *webdav.Handler

	// Rules is an array of access and deny rules.
	Rules []*Rule `json:"-"`

	// TODO: this MUST be done in another way
	StyleSheet string `json:"-"`

	// These indicate if the user can perform certain actions.
	AllowNew      bool // Create files and folders
	AllowEdit     bool // Edit/rename files
	AllowCommands bool // Execute commands

	// Commands is the list of commands the user can execute.
	Commands []string
}

// assets are the static and front-end assets, such as JS, CSS and HTML templates.
type assets struct {
	requiredJS *rice.Box // JS that is always required to have in order to be usable.
	Templates  *rice.Box
	CSS        *rice.Box
	JS         *rice.Box
}

// Rule is a dissalow/allow rule.
type Rule struct {
	// Regex indicates if this rule uses Regular Expressions or not.
	Regex bool

	// Allow indicates if this is an allow rule. Set 'false' to be a disallow rule.
	Allow bool

	// Path is the corresponding URL path for this rule.
	Path string

	// Regexp is the regular expression. Only use this when 'Regex' was set to true.
	Regexp *regexp.Regexp
}

// CommandFunc ...
type CommandFunc func(r *http.Request, c *FileManager, u *User) error

func New() *FileManager {
	m := &FileManager{
		User: &User{
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
		BeforeSave: func(r *http.Request, c *FileManager, u *User) error { return nil },
		AfterSave:  func(r *http.Request, c *FileManager, u *User) error { return nil },
		Assets: &assets{
			Templates:  rice.MustFindBox("./_assets/templates"),
			CSS:        rice.MustFindBox("./_assets/css"),
			requiredJS: rice.MustFindBox("./_assets/js"),
		},
	}

	m.SetScope(".")
	m.SetBaseURL("/")
	m.SetWebDavURL("/webdav")

	return m
}

// AbsoluteURL returns the actual URL where
// File Manager interface can be accessed.
func (m FileManager) AbsoluteURL() string {
	return m.PrefixURL + m.BaseURL
}

// AbsoluteWebDavURL returns the actual URL
// where WebDAV can be accessed.
func (m FileManager) AbsoluteWebDavURL() string {
	return m.PrefixURL + m.WebDavURL
}

// SetBaseURL updates the BaseURL of a File Manager
// object.
func (m *FileManager) SetBaseURL(url string) {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	url = "/" + url
	m.BaseURL = strings.TrimSuffix(url, "/")
}

// SetWebDavURL updates the WebDavURL of a File Manager
// object and updates it's main handler.
func (m *FileManager) SetWebDavURL(url string) {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")

	m.WebDavURL = m.BaseURL + "/" + url
	m.User.handler = &webdav.Handler{
		Prefix:     m.WebDavURL,
		FileSystem: m.fileSystem,
		LockSystem: webdav.NewMemLS(),
	}
}

// SetScope updates a user scope and its virtual file system.
func (m *FileManager) SetScope(scope string, user string) {
	m.scope = strings.TrimSuffix(scope, "/")
	m.fileSystem = webdav.Dir(m.scope)
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

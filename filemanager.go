package filemanager

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"github.com/asdine/storm"
	"golang.org/x/net/webdav"
)

var (
	// ErrDuplicated occurs when you try to create a user that already exists.
	ErrDuplicated = errors.New("Duplicated user")
)

// FileManager is a file manager instance. It should be creating using the
// 'New' function and not directly.
type FileManager struct {
	db  *storm.DB
	key []byte

	// PrefixURL is a part of the URL that is already trimmed from the request URL before it
	// arrives to our handlers. It may be useful when using File Manager as a middleware
	// such as in caddy-filemanager plugin. It is only useful in certain situations.
	PrefixURL string

	// BaseURL is the path where the GUI will be accessible. It musn't end with
	// a trailing slash and mustn't contain PrefixURL, if set. It shouldn't be
	// edited directly. Use SetBaseURL.
	BaseURL string

	// Users is a map with the different configurations for each user.
	Users map[string]*User

	assets *rice.Box
}

// Command is a command function.
type Command func(r *http.Request, m *FileManager, u *User) error

// User contains the configuration for each user. It should be created
// using NewUser on a File Manager instance.
type User struct {
	// ID is the required primary key with auto increment0
	ID int `storm:"id,increment"`

	// Username is the user username used to login.
	Username string `json:"username" storm:"index,unique"`

	// The hashed password. This never reaches the front-end because it's temporarily
	// emptied during JSON marshall.
	Password string `json:"password"`

	// FileSystem is the virtual file system the user has access.
	FileSystem webdav.Dir `json:"filesystem"`

	// Rules is an array of access and deny rules.
	Rules []*Rule `json:"rules"`

	// Costum styles for this user.
	CSS string `json:"css"`

	// These indicate if the user can perform certain actions.
	AllowNew      bool `json:"allowNew"`      // Create files and folders
	AllowEdit     bool `json:"allowEdit"`     // Edit/rename files
	AllowCommands bool `json:"allowCommands"` // Execute commands

	// Commands is the list of commands the user can execute.
	Commands []string `json:"commands"`
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
	Regexp *Regexp
}

// Regexp is a regular expression wrapper around native regexp.
type Regexp struct {
	Raw    string
	regexp *regexp.Regexp
}

// DefaultUser is used on New, when no 'base' user is provided.
var DefaultUser = User{
	Username:      "admin",
	Password:      "admin",
	AllowCommands: true,
	AllowEdit:     true,
	AllowNew:      true,
	Commands:      []string{},
	Rules:         []*Rule{},
	CSS:           "",
	FileSystem:    webdav.Dir("."),
}

// New creates a new File Manager instance. If 'database' file already
// exists, it will load the users from there. Otherwise, a new user
// will be created using the 'base' variable. The 'base' User should
// not have the Password field hashed.
// TODO: should it ask for a baseURL on New????
func New(database string, base User) (*FileManager, error) {
	// Creates a new File Manager instance with the Users
	// map and Assets box.
	m := &FileManager{
		Users:  map[string]*User{},
		assets: rice.MustFindBox("./assets/dist"),
	}

	// Tries to open a database on the location provided. This
	// function will automatically create a new one if it doesn't
	// exist.
	db, err := storm.Open(database)
	if err != nil {
		return nil, err
	}

	// Tries to get the encryption key from the database.
	// If it doesn't exist, create a new one of 256 bits.
	err = db.Get("config", "key", &m.key)
	if err != nil && err == storm.ErrNotFound {
		m.key = []byte(randomString(64))
		err = db.Set("config", "key", m.key)
	}

	if err != nil {
		return nil, err
	}

	// Tries to fetch the users from the database and if there are
	// any, add them to the current File Manager instance.
	var users []User
	err = db.All(&users)
	if err != nil {
		return nil, err
	}

	for i := range users {
		m.Users[users[i].Username] = &users[i]
	}

	// If there are no users in the database, it creates a new one
	// based on 'base' User that must be provided by the function caller.
	if len(users) == 0 {
		// Hashes the password.
		pw, err := hashPassword(base.Password)
		if err != nil {
			return nil, err
		}

		base.Password = pw

		// Saves the user to the database.
		if err := db.Save(&base); err != nil {
			return nil, err
		}

		m.Users[base.Username] = &base
	}

	// Attaches db to this File Manager instance.
	m.db = db
	return m, nil
}

// RootURL returns the actual URL where
// File Manager interface can be accessed.
func (m FileManager) RootURL() string {
	return m.PrefixURL + m.BaseURL
}

// SetPrefixURL updates the prefixURL of a File
// Manager object.
func (m *FileManager) SetPrefixURL(url string) {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	url = "/" + url
	m.PrefixURL = strings.TrimSuffix(url, "/")
}

// SetBaseURL updates the baseURL of a File Manager
// object.
func (m *FileManager) SetBaseURL(url string) {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	url = "/" + url
	m.BaseURL = strings.TrimSuffix(url, "/")
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
func (m *FileManager) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	// TODO: Handle errors here and make it compatible with http.Handler
	return serveHTTP(&requestContext{
		fm: m,
		us: nil,
		fi: nil,
	}, w, r)
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

// SetScope updates a user scope and its virtual file system.
// If the user string is blank, it will change the base scope.
func (u *User) SetScope(scope string) {
	scope = strings.TrimSuffix(scope, "/")
	u.FileSystem = webdav.Dir(scope)
}

// MatchString checks if this string matches the regular expression.
func (r *Regexp) MatchString(s string) bool {
	if r.regexp == nil {
		r.regexp = regexp.MustCompile(r.Raw)
	}

	return r.regexp.MatchString(s)
}

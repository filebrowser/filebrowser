// Package filemanager provides a web interface to access your files
// wherever you are. To use this package as a middleware for your app,
// you'll need to create a filemanager instance:
//
// 		m, err := filemanager.New(database, user)
//
// Where 'user' contains the default options for new users. You can just
// use 'filemanager.DefaultUser' or create yourself a default user:
//
// 		m, err := filemanager.New(database, filemanager.User{
// 			Admin: 		   false,
// 			AllowCommands: false,
// 			AllowEdit:     true,
// 			AllowNew:      true,
// 			Commands:      []string{
// 				"git",
// 			},
// 			Rules:         []*filemanager.Rule{},
// 			CSS:           "",
// 			FileSystem:    webdav.Dir("/path/to/files"),
// 		})
//
// The credentials for the first user are always 'admin' for both the user and
// the password, and they can be changed later through the settings. The first
// user is always an Admin and has all of the permissions set to 'true'.
//
// Then, you should set the Prefix URL and the Base URL, using the following
// functions:
//
// 		m.SetBaseURL("/")
// 		m.SetPrefixURL("/")
//
// The Prefix URL is a part of the path that is already stripped from the
// r.URL.Path variable before the request arrives to File Manager's handler.
// This is a function that will rarely be used. You can see one example on Caddy
// filemanager plugin.
//
// The Base URL is the URL path where you want File Manager to be available in. If
// you want to be available at the root path, you should call:
//
// 		m.SetBaseURL("/")
//
// But if you want to access it at '/admin', you would call:
//
// 		m.SetBaseURL("/admin")
//
// Now, that you already have a File Manager instance created, you just need to
// add it to your handlers using m.ServeHTTP which is compatible to http.Handler.
// We also have a m.ServeWithErrorsHTTP that returns the status code and an error.
//
// One simple implementation for this, at port 80, in the root of the domain, would be:
//
// 		http.ListenAndServe(":80", m)
package filemanager

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"github.com/asdine/storm"
	"github.com/hacdias/fileutils"
	"github.com/mholt/caddy"
)

var (
	errUserExist          = errors.New("user already exists")
	errUserNotExist       = errors.New("user does not exist")
	errEmptyRequest       = errors.New("request body is empty")
	errEmptyPassword      = errors.New("password is empty")
	errEmptyUsername      = errors.New("username is empty")
	errEmptyScope         = errors.New("scope is empty")
	errWrongDataType      = errors.New("wrong data type")
	errInvalidUpdateField = errors.New("invalid field to update")
	plugins               = map[string]Plugin{}
)

// FileManager is a file manager instance. It should be creating using the
// 'New' function and not directly.
type FileManager struct {
	// The BoltDB database for this instance.
	db *storm.DB

	// The key used to sign the JWT tokens.
	key []byte

	// The static assets.
	assets *rice.Box

	// PrefixURL is a part of the URL that is already trimmed from the request URL before it
	// arrives to our handlers. It may be useful when using File Manager as a middleware
	// such as in caddy-filemanager plugin. It is only useful in certain situations.
	PrefixURL string

	// BaseURL is the path where the GUI will be accessible. It musn't end with
	// a trailing slash and mustn't contain PrefixURL, if set. It shouldn't be
	// edited directly. Use SetBaseURL.
	BaseURL string

	// NoAuth disables the authentication. When the authentication is disabled,
	// there will only exist one user, called "admin".
	NoAuth bool

	// The Default User needed to build the New User page.
	DefaultUser *User

	// Users is a map with the different configurations for each user.
	Users map[string]*User

	// A map of events to a slice of commands.
	Commands map[string][]string

	// The options of the plugins that have been plugged into this instance.
	Plugins map[string]interface{}
}

// Command is a command function.
type Command func(r *http.Request, m *FileManager, u *User) error

// User contains the configuration for each user.
type User struct {
	// ID is the required primary key with auto increment0
	ID int `storm:"id,increment"`

	// Username is the user username used to login.
	Username string `json:"username" storm:"index,unique"`

	// The hashed password. This never reaches the front-end because it's temporarily
	// emptied during JSON marshall.
	Password string `json:"password"`

	// Tells if this user is an admin.
	Admin bool `json:"admin"`

	// FileSystem is the virtual file system the user has access.
	FileSystem fileutils.Dir `json:"filesystem"`

	// Rules is an array of access and deny rules.
	Rules []*Rule `json:"rules"`

	// Custom styles for this user.
	CSS string `json:"css"`

	// Locale is the language of the user.
	Locale string `json:"locale"`

	// These indicate if the user can perform certain actions.
	AllowNew      bool            `json:"allowNew"`      // Create files and folders
	AllowEdit     bool            `json:"allowEdit"`     // Edit/rename files
	AllowCommands bool            `json:"allowCommands"` // Execute commands
	Permissions   map[string]bool `json:"permissions"`   // Permissions added by plugins

	// Commands is the list of commands the user can execute.
	Commands []string `json:"commands"`
}

// Rule is a dissalow/allow rule.
type Rule struct {
	// Regex indicates if this rule uses Regular Expressions or not.
	Regex bool `json:"regex"`

	// Allow indicates if this is an allow rule. Set 'false' to be a disallow rule.
	Allow bool `json:"allow"`

	// Path is the corresponding URL path for this rule.
	Path string `json:"path"`

	// Regexp is the regular expression. Only use this when 'Regex' was set to true.
	Regexp *Regexp `json:"regexp"`
}

// Regexp is a regular expression wrapper around native regexp.
type Regexp struct {
	Raw    string `json:"raw"`
	regexp *regexp.Regexp
}

type Plugin struct {
	JavaScript    string
	CommandEvents []string
	Permissions   []Permission
	Handler       PluginHandler
	Options       interface{}
}

type Permission struct {
	Name  string
	Value bool
}

type PluginHandler interface {
	// If the Plugin returns (0, nil), the executation of File Manager will procced as usual.
	// Otherwise it will stop.
	Before(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error)
	After(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error)
}

func RegisterPlugin(name string, plugin Plugin) {
	if _, ok := plugins[name]; ok {
		panic(name + " plugin is already registred")
	}

	plugins[name] = plugin
}

// DefaultUser is used on New, when no 'base' user is provided.
var DefaultUser = User{
	AllowCommands: true,
	AllowEdit:     true,
	AllowNew:      true,
	Permissions:   map[string]bool{},
	Commands:      []string{},
	Rules:         []*Rule{},
	CSS:           "",
	Admin:         true,
	Locale:        "en",
	FileSystem:    fileutils.Dir("."),
}

// New creates a new File Manager instance. If 'database' file already
// exists, it will load the users from there. Otherwise, a new user
// will be created using the 'base' variable. The 'base' User should
// not have the Password field hashed.
func New(database string, base User) (*FileManager, error) {
	// Creates a new File Manager instance with the Users
	// map and Assets box.
	m := &FileManager{
		Users:   map[string]*User{},
		Plugins: map[string]interface{}{},
		assets:  rice.MustFindBox("./assets/dist"),
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
		var bytes []byte
		bytes, err = generateRandomBytes(64)
		if err != nil {
			return nil, err
		}

		m.key = bytes
		err = db.Set("config", "key", m.key)
	}

	if err != nil {
		return nil, err
	}

	// Tries to get the event commands from the database.
	// If they don't exist, initialize them.
	err = db.Get("config", "commands", &m.Commands)
	if err != nil && err == storm.ErrNotFound {
		m.Commands = map[string][]string{
			"before_save": {},
			"after_save":  {},
		}
		err = db.Set("config", "commands", m.Commands)
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
		u := base
		u.Username = "admin"

		// Hashes the password.
		u.Password, err = hashPassword("admin")
		if err != nil {
			return nil, err
		}

		// The first user must be an administrator.
		u.Admin = true
		u.AllowCommands = true
		u.AllowNew = true
		u.AllowEdit = true

		// Saves the user to the database.
		if err := db.Save(&u); err != nil {
			return nil, err
		}

		m.Users[u.Username] = &u
	}

	// Attaches db to this File Manager instance.
	m.db = db

	// Create the default user, making a copy of the base.
	base.Username = ""
	base.Password = ""
	m.DefaultUser = &base
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

// ActivatePlugin activates a plugin to a File Manager instance and
// loads its options from the database.
func (m *FileManager) ActivatePlugin(name string, options interface{}) error {
	if reflect.TypeOf(options).Kind() != reflect.Ptr {
		return errors.New("options should be a pointer to interface, not interface")
	}

	var plugin Plugin

	if p, ok := plugins[name]; !ok {
		plugin = p
		return errors.New(name + " plugin is not registred")
	}

	if _, ok := m.Plugins[name]; ok {
		return errors.New(name + " plugin is already activated")
	}

	err := m.db.Get("plugins", name, &plugin)
	if err != nil && err == storm.ErrNotFound {
		err = m.db.Set("plugin", name, plugin)
	}

	if err != nil {
		return err
	}

	// Register the command event hooks.
	for _, evt := range plugin.CommandEvents {
		if _, ok := m.Commands[evt]; ok {
			continue
		}

		m.Commands[evt] = []string{}
	}

	err = m.db.Set("config", "commands", m.Commands)
	if err != nil {
		return err
	}

	// Register the user permissions.
	for _, perm := range plugin.Permissions {
		err = m.registerPermission(perm.Name, perm.Value)
		if err != nil {
			return err
		}
	}

	m.Plugins[name] = options
	return nil
}

// registerPermission registers a new user permission and adds it to every
// user with it default's 'value'. If the user is an admin, it will
// be true.
func (m *FileManager) registerPermission(name string, value bool) error {
	if _, ok := m.DefaultUser.Permissions[name]; ok {
		return nil
	}

	// Add the default value for this permission on the default user.
	m.DefaultUser.Permissions[name] = value

	for _, u := range m.Users {
		// Bypass the user if it is already defined.
		if _, ok := u.Permissions[name]; ok {
			continue
		}

		if u.Permissions == nil {
			u.Permissions = m.DefaultUser.Permissions
		}

		if u.Admin {
			u.Permissions[name] = true
		}

		err := m.db.Save(u)
		if err != nil {
			return err
		}
	}

	return nil
}

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
// Compatible with http.Handler.
func (m *FileManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, err := serveHTTP(&RequestContext{
		FileManager: m,
		User:        nil,
		File:        nil,
	}, w, r)

	if code >= 400 {
		w.WriteHeader(code)

		if err == nil {
			txt := http.StatusText(code)
			log.Printf("%v: %v %v\n", r.URL.Path, code, txt)
			w.Write([]byte(txt))
		}
	}

	if err != nil {
		log.Print(err)
		w.Write([]byte(err.Error()))
	}
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

// MatchString checks if this string matches the regular expression.
func (r *Regexp) MatchString(s string) bool {
	if r.regexp == nil {
		r.regexp = regexp.MustCompile(r.Raw)
	}

	return r.regexp.MatchString(s)
}

// Runner runs the commands for a certain event type.
func (m FileManager) Runner(event string, path string) error {
	commands := []string{}

	// Get the commands from the File Manager instance itself.
	if val, ok := m.Commands[event]; ok {
		commands = append(commands, val...)
	}

	// Execute the commands.
	for _, command := range commands {
		args := strings.Split(command, " ")
		nonblock := false

		if len(args) > 1 && args[len(args)-1] == "&" {
			// Run command in background; non-blocking
			nonblock = true
			args = args[:len(args)-1]
		}

		command, args, err := caddy.SplitCommandAndArgs(strings.Join(args, " "))
		if err != nil {
			return err
		}

		cmd := exec.Command(command, args...)
		cmd.Env = append(os.Environ(), "file="+path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if nonblock {
			log.Printf("[INFO] Nonblocking Command:\"%s %s\"", command, strings.Join(args, " "))
			if err := cmd.Start(); err != nil {
				return err
			}

			continue
		}

		log.Printf("[INFO] Blocking Command:\"%s %s\"", command, strings.Join(args, " "))
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

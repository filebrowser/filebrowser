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
	"crypto/rand"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	rice "github.com/GeertJohan/go.rice"
	"github.com/asdine/storm"
	"github.com/hacdias/fileutils"
	"github.com/mholt/caddy"
	"github.com/robfig/cron"
)

var (
	ErrExist              = errors.New("the resource already exists")
	ErrNotExist           = errors.New("the resource does not exist")
	ErrEmptyRequest       = errors.New("request body is empty")
	ErrEmptyPassword      = errors.New("password is empty")
	ErrEmptyUsername      = errors.New("username is empty")
	ErrEmptyScope         = errors.New("scope is empty")
	ErrWrongDataType      = errors.New("wrong data type")
	ErrInvalidUpdateField = errors.New("invalid field to update")
	ErrInvalidOption      = errors.New("Invalid option")
)

// FileManager is a file manager instance. It should be creating using the
// 'New' function and not directly.
type FileManager struct {
	// Job cron.
	Cron *cron.Cron

	// The key used to sign the JWT tokens.
	Key []byte

	// The static assets.
	Assets *rice.Box

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

	// StaticGen is the static websit generator handler.
	StaticGen StaticGen

	// The Default User needed to build the New User page.
	DefaultUser *User

	// A map of events to a slice of commands.
	Commands map[string][]string

	Store *Store
}

// Command is a command function.
type Command func(r *http.Request, m *FileManager, u *User) error

// Load loads the configuration from the database.
func (m *FileManager) Load() error {
	// Creates a new File Manager instance with the Users
	// map and Assets box.
	m.Assets = rice.MustFindBox("./assets/dist")
	m.Cron = cron.New()

	// Tries to get the encryption key from the database.
	// If it doesn't exist, create a new one of 256 bits.
	err := m.Store.Config.Get("key", &m.Key)
	if err != nil && err == ErrNotExist {
		var bytes []byte
		bytes, err = GenerateRandomBytes(64)
		if err != nil {
			return err
		}

		m.Key = bytes
		err = m.Store.Config.Save("key", m.Key)
	}

	if err != nil {
		return err
	}

	// Tries to get the event commands from the database.
	// If they don't exist, initialize them.
	err = m.Store.Config.Get("commands", &m.Commands)
	if err != nil && err == storm.ErrNotFound {
		m.Commands = map[string][]string{
			"before_save":    {},
			"after_save":     {},
			"before_publish": {},
			"after_publish":  {},
		}
		err = m.Store.Config.Save("commands", m.Commands)
	}

	if err != nil {
		return err
	}

	// Tries to fetch the users from the database.
	users, err := m.Store.Users.Gets()
	if err != nil {
		return err
	}

	// If there are no users in the database, it creates a new one
	// based on 'base' User that must be provided by the function caller.
	if len(users) == 0 {
		u := *m.DefaultUser
		u.Username = "admin"

		// Hashes the password.
		u.Password, err = HashPassword("admin")
		if err != nil {
			return err
		}

		// The first user must be an administrator.
		u.Admin = true
		u.AllowCommands = true
		u.AllowNew = true
		u.AllowEdit = true
		u.AllowPublish = true

		// Saves the user to the database.
		if err := m.Store.Users.Save(&u); err != nil {
			return err
		}
	}

	m.DefaultUser.Username = ""
	m.DefaultUser.Password = ""

	m.Cron.AddFunc("@hourly", m.ShareCleaner)
	m.Cron.Start()

	return nil
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

// Attach attaches a static generator to the current File Manager.
func (m *FileManager) Attach(s StaticGen) error {
	if reflect.TypeOf(s).Kind() != reflect.Ptr {
		return errors.New("data should be a pointer to interface, not interface")
	}

	err := s.Setup()
	if err != nil {
		return err
	}

	m.StaticGen = s

	err = m.Store.Config.Get("staticgen_"+s.Name(), s)
	if err == ErrNotExist {
		return m.Store.Config.Save("staticgen_"+s.Name(), s)
	}

	return err
}

// ShareCleaner removes sharing links that are no longer active.
// This function is set to run periodically.
func (m FileManager) ShareCleaner() {
	// Get all links.
	links, err := m.Store.Share.Gets()
	if err != nil {
		log.Print(err)
		return
	}

	// Find the expired ones.
	for i := range links {
		if links[i].Expires && links[i].ExpireDate.Before(time.Now()) {
			err = m.Store.Share.Delete(links[i].Hash)
			if err != nil {
				log.Print(err)
			}
		}
	}
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

// DefaultUser is used on New, when no 'base' user is provided.
var DefaultUser = User{
	AllowCommands: true,
	AllowEdit:     true,
	AllowNew:      true,
	AllowPublish:  true,
	Commands:      []string{},
	Rules:         []*Rule{},
	CSS:           "",
	Admin:         true,
	Locale:        "en",
	FileSystem:    fileutils.Dir("."),
}

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
	AllowNew      bool `json:"allowNew"`      // Create files and folders
	AllowEdit     bool `json:"allowEdit"`     // Edit/rename files
	AllowCommands bool `json:"allowCommands"` // Execute commands
	AllowPublish  bool `json:"allowPublish"`  // Publish content (to use with static gen)

	// Commands is the list of commands the user can execute.
	Commands []string `json:"commands"`
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

// MatchString checks if this string matches the regular expression.
func (r *Regexp) MatchString(s string) bool {
	if r.regexp == nil {
		r.regexp = regexp.MustCompile(r.Raw)
	}

	return r.regexp.MatchString(s)
}

// ShareLink is the information needed to build a shareable link.
type ShareLink struct {
	Hash       string    `json:"hash" storm:"id,index"`
	Path       string    `json:"path" storm:"index"`
	Expires    bool      `json:"expires"`
	ExpireDate time.Time `json:"expireDate"`
}

// Store is a collection of the stores needed to get
// and save information.
type Store struct {
	Users  UsersStore
	Config ConfigStore
	Share  ShareStore
}

// UsersStore is the interface to manage users.
type UsersStore interface {
	Get(id int) (*User, error)
	Gets() ([]*User, error)
	Save(u *User) error
	Update(u *User, fields ...string) error
	Delete(id int) error
}

// ConfigStore is the interface to manage configuration.
type ConfigStore interface {
	Get(name string, to interface{}) error
	Save(name string, from interface{}) error
}

// ShareStore is the interface to manage share links.
type ShareStore interface {
	Get(hash string) (*ShareLink, error)
	GetPermanent(path string) (*ShareLink, error)
	GetByPath(path string) ([]*ShareLink, error)
	Gets() ([]*ShareLink, error)
	Save(s *ShareLink) error
	Delete(hash string) error
}

// StaticGen is a static website generator.
type StaticGen interface {
	SettingsPath() string
	Name() string
	Setup() error

	Hook(c *Context, w http.ResponseWriter, r *http.Request) (int, error)
	Preview(c *Context, w http.ResponseWriter, r *http.Request) (int, error)
	Publish(c *Context, w http.ResponseWriter, r *http.Request) (int, error)
}

// Context contains the needed information to make handlers work.
type Context struct {
	*FileManager
	User *User
	File *File
	// On API handlers, Router is the APi handler we want.
	Router string
}

// HashPassword generates an hash from a password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with an hash to check if they match.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an fm.Error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

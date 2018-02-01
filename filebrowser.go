package filebrowser

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/GeertJohan/go.rice"
	"github.com/hacdias/fileutils"
	"github.com/mholt/caddy"
	"github.com/robfig/cron"
)

const (
	// Version is the current File Manager version.
	Version = "1.5.0"

	ListViewMode   = "list"
	MosaicViewMode = "mosaic"
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
	ErrInvalidOption      = errors.New("invalid option")
)

// FileBrowser is a file manager instance. It should be creating using the
// 'New' function and not directly.
type FileBrowser struct {
	// Cron job to manage schedulings.
	Cron *cron.Cron

	// The key used to sign the JWT tokens.
	Key []byte

	// The static assets.
	Assets *rice.Box

	// The Store is used to manage users, shareable links and
	// other stuff that is saved on the database.
	Store *Store

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

	// ReCaptcha Site key and secret.
	ReCaptchaKey    string
	ReCaptchaSecret string

	// StaticGen is the static websit generator handler.
	StaticGen StaticGen

	// The Default User needed to build the New User page.
	DefaultUser *User

	// A map of events to a slice of commands.
	Commands map[string][]string

	// Global stylesheet.
	CSS string

	// NewFS should build a new file system for a given path.
	NewFS FSBuilder
}

var commandEvents = []string{
	"before_save",
	"after_save",
	"before_publish",
	"after_publish",
	"before_copy",
	"after_copy",
	"before_rename",
	"after_rename",
	"before_upload",
	"after_upload",
	"before_delete",
	"after_delete",
}

// Command is a command function.
type Command func(r *http.Request, m *FileBrowser, u *User) error

// FSBuilder is the File System Builder.
type FSBuilder func(scope string) FileSystem

// Setup loads the configuration from the database and configures
// the Assets and the Cron job. It must always be run after
// creating a File Manager object.
func (m *FileBrowser) Setup() error {
	// Creates a new File Manager instance with the Users
	// map and Assets box.
	m.Assets = rice.MustFindBox("./node_modules/filebrowser-frontend/dist")
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

	// Get the global CSS.
	err = m.Store.Config.Get("css", &m.CSS)
	if err != nil && err == ErrNotExist {
		err = m.Store.Config.Save("css", "")
	}

	if err != nil {
		return err
	}

	// Tries to get the event commands from the database.
	// If they don't exist, initialize them.
	err = m.Store.Config.Get("commands", &m.Commands)

	if err == nil {
		// Add hypothetically new command handlers.
		for _, command := range commandEvents {
			if _, ok := m.Commands[command]; ok {
				continue
			}

			m.Commands[command] = []string{}
		}
	}

	if err != nil && err == ErrNotExist {
		m.Commands = map[string][]string{}

		// Initialize the command handlers.
		for _, command := range commandEvents {
			m.Commands[command] = []string{}
		}

		err = m.Store.Config.Save("commands", m.Commands)
	}

	if err != nil {
		return err
	}

	// Tries to fetch the users from the database.
	users, err := m.Store.Users.Gets(m.NewFS)
	if err != nil && err != ErrNotExist {
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
func (m FileBrowser) RootURL() string {
	return m.PrefixURL + m.BaseURL
}

// SetPrefixURL updates the prefixURL of a File
// Manager object.
func (m *FileBrowser) SetPrefixURL(url string) {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	url = "/" + url
	m.PrefixURL = strings.TrimSuffix(url, "/")
}

// SetBaseURL updates the baseURL of a File Manager
// object.
func (m *FileBrowser) SetBaseURL(url string) {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	url = "/" + url
	m.BaseURL = strings.TrimSuffix(url, "/")
}

// Attach attaches a static generator to the current File Manager.
func (m *FileBrowser) Attach(s StaticGen) error {
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
func (m FileBrowser) ShareCleaner() {
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
func (m FileBrowser) Runner(event string, path string, destination string, user *User) error {
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
		cmd.Env = append(os.Environ(), fmt.Sprintf("FILE=%s", path))
		cmd.Env = append(cmd.Env, fmt.Sprintf("ROOT=%s", user.Scope))
		cmd.Env = append(cmd.Env, fmt.Sprintf("TRIGGER=%s", event))
		cmd.Env = append(cmd.Env, fmt.Sprintf("USERNAME=%s", user.Username))

		if destination != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("DESTINATION=%s", destination))
		}

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
	LockPassword:  false,
	Commands:      []string{},
	Rules:         []*Rule{},
	CSS:           "",
	Admin:         true,
	Locale:        "",
	Scope:         ".",
	FileSystem:    fileutils.Dir("."),
	ViewMode:      "mosaic",
}

// User contains the configuration for each user.
type User struct {
	// ID is the required primary key with auto increment0
	ID int `storm:"id,increment"`

	// Tells if this user is an admin.
	Admin bool `json:"admin"`

	// These indicate if the user can perform certain actions.
	AllowCommands bool `json:"allowCommands"` // Execute commands
	AllowEdit     bool `json:"allowEdit"`     // Edit/rename files
	AllowNew      bool `json:"allowNew"`      // Create files and folders
	AllowPublish  bool `json:"allowPublish"`  // Publish content (to use with static gen)

	// Prevents the user to change its password.
	LockPassword bool `json:"lockPassword"`

	// Commands is the list of commands the user can execute.
	Commands []string `json:"commands"`

	// Custom styles for this user.
	CSS string `json:"css"`

	// FileSystem is the virtual file system the user has access.
	FileSystem FileSystem `json:"-"`

	// Locale is the language of the user.
	Locale string `json:"locale"`

	// The hashed password. This never reaches the front-end because it's temporarily
	// emptied during JSON marshall.
	Password string `json:"password"`

	// Rules is an array of access and deny rules.
	Rules []*Rule `json:"rules"`

	// Scope is the path the user has access to.
	Scope string `json:"filesystem"`

	// Username is the user username used to login.
	Username string `json:"username" storm:"index,unique"`

	// User view mode for files and folders.
	ViewMode string `json:"viewMode"`
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
	Get(id int, builder FSBuilder) (*User, error)
	GetByUsername(username string, builder FSBuilder) (*User, error)
	Gets(builder FSBuilder) ([]*User, error)
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

// FileSystem is the interface to work with the file system.
type FileSystem interface {
	Mkdir(name string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	RemoveAll(name string) error
	Rename(oldName, newName string) error
	Stat(name string) (os.FileInfo, error)
	Copy(src, dst string) error
}

// Context contains the needed information to make handlers work.
type Context struct {
	*FileBrowser
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

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
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/asdine/storm"
	"github.com/hacdias/fileutils"
	"github.com/mholt/caddy"
	"github.com/robfig/cron"
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

	// Job cron.
	cron *cron.Cron

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

	// staticgen is the name of the current static website generator.
	staticgen string
	// StaticGen is the static websit generator handler.
	StaticGen StaticGen

	// The Default User needed to build the New User page.
	DefaultUser *User

	// Users is a map with the different configurations for each user.
	Users map[string]*User

	// A map of events to a slice of commands.
	Commands map[string][]string
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
	AllowNew      bool `json:"allowNew"`      // Create files and folders
	AllowEdit     bool `json:"allowEdit"`     // Edit/rename files
	AllowCommands bool `json:"allowCommands"` // Execute commands
	AllowPublish  bool `json:"allowPublish"`  // Publish content (to use with static gen)

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

// New creates a new File Manager instance. If 'database' file already
// exists, it will load the users from there. Otherwise, a new user
// will be created using the 'base' variable. The 'base' User should
// not have the Password field hashed.
func New(database string, base User) (*FileManager, error) {
	// Creates a new File Manager instance with the Users
	// map and Assets box.
	m := &FileManager{
		Users:  map[string]*User{},
		cron:   cron.New(),
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
			"before_save":    {},
			"after_save":     {},
			"before_publish": {},
			"after_publish":  {},
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
		u.AllowPublish = true

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

	m.cron.AddFunc("@hourly", m.shareCleaner)
	m.cron.Start()

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

// ServeHTTP handles the request.
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

// EnableStaticGen attaches a static generator to the current File Manager
// instance.
func (m *FileManager) EnableStaticGen(data StaticGen) error {
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return errors.New("data should be a pointer to interface, not interface")
	}

	if h, ok := data.(*Hugo); ok {
		return m.enableHugo(h)
	}

	if j, ok := data.(*Jekyll); ok {
		return m.enableJekyll(j)
	}

	return errors.New("unknown static website generator")
}

func (m *FileManager) enableHugo(h *Hugo) error {
	if err := h.find(); err != nil {
		return err
	}

	m.staticgen = "hugo"
	m.StaticGen = h

	err := m.db.Get("staticgen", "hugo", h)
	if err != nil && err == storm.ErrNotFound {
		err = m.db.Set("staticgen", "hugo", *h)
	}

	return nil
}

func (m *FileManager) enableJekyll(j *Jekyll) error {
	if err := j.find(); err != nil {
		return err
	}

	if len(j.Args) == 0 {
		j.Args = []string{"build"}
	}

	if j.Args[0] != "build" {
		j.Args = append([]string{"build"}, j.Args...)
	}

	m.staticgen = "jekyll"
	m.StaticGen = j

	err := m.db.Get("staticgen", "jekyll", j)
	if err != nil && err == storm.ErrNotFound {
		err = m.db.Set("staticgen", "jekyll", *j)
	}

	return nil
}

// shareCleaner removes sharing links that are no longer active.
// This function is set to run periodically.
func (m FileManager) shareCleaner() {
	var links []shareLink

	// Get all links.
	err := m.db.All(&links)
	if err != nil {
		log.Print(err)
		return
	}

	// Find the expired ones.
	for i := range links {
		if links[i].Expires && links[i].ExpireDate.Before(time.Now()) {
			err = m.db.DeleteStruct(&links[i])
			if err != nil {
				log.Print(err)
			}
		}
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

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
	"strings"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/asdine/storm"
	"github.com/mholt/caddy"
	"github.com/robfig/cron"
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

// ServeHTTP handles the request.
func (m *FileManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	/* code, err := serveHTTP(&RequestContext{
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
	} */
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

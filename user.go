package filemanager

import (
	"crypto/rand"
	"errors"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/hacdias/fileutils"
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
)

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

type ShareLink struct {
	Hash       string    `json:"hash" storm:"id,index"`
	Path       string    `json:"path" storm:"index"`
	Expires    bool      `json:"expires"`
	ExpireDate time.Time `json:"expireDate"`
}

type Store struct {
	Users  UsersStore
	Config ConfigStore
	Share  ShareStore
}

type UsersStore interface {
	Get(id int) (*User, error)
	Gets() ([]*User, error)
	Save(u *User) error
	Update(u *User, fields ...string) error
	Delete(id int) error
}

type ConfigStore interface {
	Get(name string, to interface{}) error
	Save(name string, from interface{}) error
}

type ShareStore interface {
	Get(hash string) (*ShareLink, error)
	GetByPath(path string) ([]*ShareLink, error)
	Gets() ([]*ShareLink, error)
	Save(s *ShareLink) error
	Delete(hash string) error
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

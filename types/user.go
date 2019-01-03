package types

import (
	"path/filepath"
	"regexp"

	"github.com/spf13/afero"
	"golang.org/x/crypto/bcrypt"
)

// ViewMode describes a view mode.
type ViewMode string

const (
	ListViewMode   ViewMode = "list"
	MosaicViewMode ViewMode = "mosaic"
)

// Permissions describe a user's permissions.
type Permissions struct {
	Admin    bool `json:"admin"`
	Execute  bool `json:"execute"`
	Create   bool `json:"create"`
	Rename   bool `json:"rename"`
	Modify   bool `json:"edit"`
	Delete   bool `json:"delete"`
	Share    bool `json:"share"`
	Download bool `json:"download"`
}

// User describes a user.
type User struct {
	ID           uint        `storm:"id,increment" json:"id"`
	Username     string      `storm:"unique" json:"username"`
	Password     string      `json:"password"`
	Scope        string      `json:"scope"`
	Locale       string      `json:"locale"`
	LockPassword bool        `json:"lockPassword"`
	ViewMode     ViewMode    `json:"viewMode"`
	Perm         Permissions `json:"perm"`
	Commands     []string    `json:"commands"`
	Sorting      Sorting     `json:"sorting"`
	Fs           afero.Fs    `json:"-"`
	Rules        []Rule      `json:"rules"`
}

var checkableFields = []string{
	"Username",
	"Password",
	"Scope",
	"ViewMode",
	"Commands",
	"Sorting",
	"Rules",
}

func (u *User) clean(settings *Settings, fields ...string) error {
	if len(fields) == 0 {
		fields = checkableFields
	}

	for _, field := range fields {
		switch field {
		case "Username":
			if u.Username == "" {
				return ErrEmptyUsername
			}
		case "Password":
			if u.Password == "" {
				return ErrEmptyPassword
			}
		case "Scope":
			if !filepath.IsAbs(u.Scope) {
				return ErrPathIsRel
			}
		case "ViewMode":
			if u.ViewMode == "" {
				u.ViewMode = ListViewMode
			}
		case "Commands":
			if u.Commands == nil {
				u.Commands = []string{}
			}
		case "Sorting":
			if u.Sorting.By == "" {
				u.Sorting.By = "name"
			}
		case "Rules":
			if u.Rules == nil {
				u.Rules = []Rule{}
			}
		}
	}

	if u.Fs == nil {
		u.Fs = &userFs{
			user:     u,
			settings: settings,
			source:   afero.NewBasePathFs(afero.NewOsFs(), u.Scope),
		}
	}

	return nil
}

// CanExecute checks if an user can execute a specific command.
func (u *User) CanExecute(command string) bool {
	if !u.Perm.Execute {
		return false
	}

	for _, cmd := range u.Commands {
		if regexp.MustCompile(cmd).MatchString(command) {
			return true
		}
	}

	return false
}

// ApplyDefaults applies defaults to a user.
func (u *User) ApplyDefaults(defaults UserDefaults) {
	u.Scope = defaults.Scope
	u.Locale = defaults.Locale
	u.ViewMode = defaults.ViewMode
	u.Perm = defaults.Perm
	u.Sorting = defaults.Sorting
	u.Commands = defaults.Commands
}

// HashPwd hashes a password.
func HashPwd(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPwd checks if a password is correct.
func CheckPwd(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

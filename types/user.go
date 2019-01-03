package types

import (
	"path/filepath"
	"regexp"

	"github.com/spf13/afero"
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
	Modify   bool `json:"modify"`
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

func (u *User) clean(fields ...string) error {
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
		u.Fs = afero.NewBasePathFs(afero.NewOsFs(), u.Scope)
	}

	return nil
}

// FullPath gets the full path for a user's relative path.
func (u *User) FullPath(path string) string {
	return afero.FullBaseFsPath(u.Fs.(*afero.BasePathFs), path)
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

package users

import (
	"path/filepath"
	"regexp"

	"github.com/filebrowser/filebrowser/v2/errors"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/spf13/afero"
)

// ViewMode describes a view mode.
type ViewMode string

const (
	ListViewMode   ViewMode = "list"
	MosaicViewMode ViewMode = "mosaic"
)

// User describes a user.
type User struct {
	ID           uint          `storm:"id,increment" json:"id"`
	Username     string        `storm:"unique" json:"username"`
	Password     string        `json:"password"`
	Scope        string        `json:"scope"`
	Locale       string        `json:"locale"`
	LockPassword bool          `json:"lockPassword"`
	ViewMode     ViewMode      `json:"viewMode"`
	Perm         Permissions   `json:"perm"`
	Commands     []string      `json:"commands"`
	Sorting      files.Sorting `json:"sorting"`
	Fs           afero.Fs      `json:"-" yaml:"-"`
	Rules        []rules.Rule  `json:"rules"`
	Bookmarks    []Bookmark    `json:"bookmarks"`
}

// GetRules implements rules.Provider.
func (u *User) GetRules() []rules.Rule {
	return u.Rules
}

var checkableFields = []string{
	"Username",
	"Password",
	"Scope",
	"ViewMode",
	"Commands",
	"Sorting",
	"Rules",
	"Bookmarks",
}

// Clean cleans up a user and verifies if all its fields
// are alright to be saved.
func (u *User) Clean(baseScope string, fields ...string) error {
	if len(fields) == 0 {
		fields = checkableFields
	}

	for _, field := range fields {
		switch field {
		case "Username":
			if u.Username == "" {
				return errors.ErrEmptyUsername
			}
		case "Password":
			if u.Password == "" {
				return errors.ErrEmptyPassword
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
				u.Rules = []rules.Rule{}
			}
		case "Bookmarks":
			if u.Bookmarks == nil {
				u.Bookmarks = []Bookmark{}
			}
		}
	}

	if u.Fs == nil {
		scope := u.Scope

		if !filepath.IsAbs(scope) {
			scope = filepath.Join(baseScope, scope)
		}

		u.Fs = afero.NewBasePathFs(afero.NewOsFs(), scope)
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

// RemoveBookmarkByPath removes a bookmark by path.
func (u *User) RemoveBookmarkByPath(path string) {
	var newBookmarks []Bookmark
	for _, bookmark := range u.Bookmarks {
		if bookmark.Path != path {
			newBookmarks = append(newBookmarks, bookmark)
		}
	}
	u.Bookmarks = newBookmarks
}

// AddBookmark adds a bookmark.
func (u *User) AddBookmark(path, name string) {
	for i, bookmark := range u.Bookmarks {
		if bookmark.Path == path {
			if name != "" {
				u.Bookmarks[i].Name = name
			}
			return
		}
	}

	if name == "" {
		name = filepath.Base(path)
	}

	u.Bookmarks = append(u.Bookmarks, Bookmark{
		Path: path,
		Name: name,
	})
}

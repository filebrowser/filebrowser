package settings

import (
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/users"
)

// UserDefaults is a type that holds the default values
// for some fields on User.
type UserDefaults struct {
	Scope    string            `json:"scope"`
	Locale   string            `json:"locale"`
	ViewMode users.ViewMode    `json:"viewMode"`
	Sorting  files.Sorting     `json:"sorting"`
	Perm     users.Permissions `json:"perm"`
	Commands []string          `json:"commands"`
}

// Apply applies the default options to a user.
func (d *UserDefaults) Apply(u *users.User) {
	u.Scope = d.Scope
	u.Locale = d.Locale
	u.ViewMode = d.ViewMode
	u.Perm = d.Perm
	u.Sorting = d.Sorting
	u.Commands = d.Commands
}

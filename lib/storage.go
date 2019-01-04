package lib

import (
	"strings"
)

// StorageBackend is the interface used to persist data.
type StorageBackend interface {
	GetUserByID(uint) (*User, error)
	GetUserByUsername(string) (*User, error)
	GetUsers() ([]*User, error)
	SaveUser(u *User) error
	UpdateUser(u *User, fields ...string) error
	DeleteUserByID(uint) error
	DeleteUserByUsername(string) error
	GetSettings() (*Settings, error)
	SaveSettings(*Settings) error
	GetAuther(AuthMethod) (Auther, error)
	SaveAuther(Auther) error
	GetLinkByHash(hash string) (*ShareLink, error)
	GetLinkPermanent(path string) (*ShareLink, error)
	GetLinksByPath(path string) ([]*ShareLink, error)
	SaveLink(s *ShareLink) error
	DeleteLink(hash string) error
}

// GetUser allows you to get a user by its name or username. The provided
// id must be a string for username lookup or a uint for id lookup. If id
// is neither, a ErrInvalidDataType will be returned.
func (f *FileBrowser) GetUser(id interface{}) (*User, error) {
	var (
		user *User
		err  error
	)

	switch id.(type) {
	case string:
		user, err = f.storage.GetUserByUsername(id.(string))
	case uint:
		user, err = f.storage.GetUserByID(id.(uint))
	default:
		return nil, ErrInvalidDataType
	}

	if err != nil {
		return nil, err
	}

	user.clean()
	return user, err
}

// GetUsers gets a list of all users.
func (f *FileBrowser) GetUsers() ([]*User, error) {
	users, err := f.storage.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		user.clean()
	}

	return users, err
}

// UpdateUser updates a user in the database.
func (f *FileBrowser) UpdateUser(user *User, fields ...string) error {
	err := user.clean(fields...)
	if err != nil {
		return err
	}

	return f.storage.UpdateUser(user, fields...)
}

// SaveUser saves the user in a storage.
func (f *FileBrowser) SaveUser(user *User) error {
	if err := user.clean(); err != nil {
		return err
	}

	return f.storage.SaveUser(user)
}

// DeleteUser allows you to delete a user by its name or username. The provided
// id must be a string for username lookup or a uint for id lookup. If id
// is neither, a ErrInvalidDataType will be returned.
func (f *FileBrowser) DeleteUser(id interface{}) (err error) {
	switch id.(type) {
	case string:
		err = f.storage.DeleteUserByUsername(id.(string))
	case uint:
		err = f.storage.DeleteUserByID(id.(uint))
	default:
		err = ErrInvalidDataType
	}

	return
}

// GetSettings returns the settings for the current instance.
func (f *FileBrowser) GetSettings() *Settings {
	return f.settings
}

// SaveSettings saves the settings for the current instance.
func (f *FileBrowser) SaveSettings(s *Settings) error {
	s.BaseURL = strings.TrimSuffix(s.BaseURL, "/")

	if len(s.Key) == 0 {
		return ErrEmptyKey
	}

	if s.Defaults.Locale == "" {
		s.Defaults.Locale = "en"
	}

	if s.Defaults.Commands == nil {
		s.Defaults.Commands = []string{}
	}

	if s.Defaults.ViewMode == "" {
		s.Defaults.ViewMode = MosaicViewMode
	}

	if s.Rules == nil {
		s.Rules = []Rule{}
	}

	if s.Shell == nil {
		s.Shell = []string{}
	}

	if s.Commands == nil {
		s.Commands = map[string][]string{}
	}

	for _, event := range defaultEvents {
		if _, ok := s.Commands["before_"+event]; !ok {
			s.Commands["before_"+event] = []string{}
		}

		if _, ok := s.Commands["after_"+event]; !ok {
			s.Commands["after_"+event] = []string{}
		}
	}

	err := f.storage.SaveSettings(s)
	if err != nil {
		return err
	}

	f.mux.Lock()
	f.settings = s
	f.mux.Unlock()
	return nil
}

// GetAuther wraps a StorageBackend.GetAuther and calls SetInstance on the auther.
func (f *FileBrowser) GetAuther(t AuthMethod) (Auther, error) {
	auther, err := f.storage.GetAuther(t)
	if err != nil {
		return nil, err
	}

	auther.SetInstance(f)
	return auther, nil
}

// SaveAuther wraps a StorageBackend.SaveAuther.
func (f *FileBrowser) SaveAuther(a Auther) error {
	return f.storage.SaveAuther(a)
}

// GetLinkByHash wraps a StorageBackend.GetLinkByHash.
func (f *FileBrowser) GetLinkByHash(hash string) (*ShareLink, error) {
	return f.storage.GetLinkByHash(hash)
}

// GetLinkPermanent wraps a StorageBackend.GetLinkPermanent
func (f *FileBrowser) GetLinkPermanent(path string) (*ShareLink, error) {
	return f.storage.GetLinkPermanent(path)
}

// GetLinksByPath wraps a StorageBackend.GetLinksByPath
func (f *FileBrowser) GetLinksByPath(path string) ([]*ShareLink, error) {
	return f.storage.GetLinksByPath(path)
}

// SaveLink wraps a StorageBackend.SaveLink
func (f *FileBrowser) SaveLink(s *ShareLink) error {
	return f.storage.SaveLink(s)
}

// DeleteLink wraps a StorageBackend.DeleteLink
func (f *FileBrowser) DeleteLink(hash string) error {
	return f.storage.DeleteLink(hash)
}

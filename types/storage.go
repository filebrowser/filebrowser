package types

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

// Storage implements Storage interface and verifies
// the data before getting in and out the database.
type Storage struct {
	src StorageBackend
}

// NewStorage creates a Storage from a StorageBackend.
func NewStorage(src StorageBackend) *Storage {
	return &Storage{src: src}
}

// GetUser allows you to get a user by its name or username. The provided
// id must be a string for username lookup or a uint for id lookup. If id
// is neither, a ErrInvalidDataType will be returned.
func (v *Storage) GetUser(id interface{}) (*User, error) {
	var (
		user *User
		err  error
	)

	switch id.(type) {
	case string:
		user, err = v.src.GetUserByUsername(id.(string))
	case uint:
		user, err = v.src.GetUserByID(id.(uint))
	default:
		return nil, ErrInvalidDataType
	}

	if err != nil {
		return nil, err
	}

	settings, err := v.GetSettings()
	if err != nil {
		return nil, err
	}

	user.clean(settings)
	return user, err
}

// GetUsers gets a list of all users.
func (v *Storage) GetUsers() ([]*User, error) {
	users, err := v.src.GetUsers()
	if err != nil {
		return nil, err
	}

	settings, err := v.GetSettings()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		user.clean(settings)
	}

	return users, err
}

// UpdateUser updates a user in the database.
func (v *Storage) UpdateUser(user *User, fields ...string) error {
	settings, err := v.GetSettings()
	if err != nil {
		return err
	}

	err = user.clean(settings, fields...)
	if err != nil {
		return err
	}

	return v.src.UpdateUser(user, fields...)
}

// SaveUser saves the user in a storage.
func (v *Storage) SaveUser(user *User) error {
	settings, err := v.GetSettings()
	if err != nil {
		return err
	}

	if err := user.clean(settings); err != nil {
		return err
	}

	return v.src.SaveUser(user)
}

// DeleteUser allows you to delete a user by its name or username. The provided
// id must be a string for username lookup or a uint for id lookup. If id
// is neither, a ErrInvalidDataType will be returned.
func (v *Storage) DeleteUser(id interface{}) (err error) {
	switch id.(type) {
	case string:
		err = v.src.DeleteUserByUsername(id.(string))
	case uint:
		err = v.src.DeleteUserByID(id.(uint))
	default:
		err = ErrInvalidDataType
	}

	return
}

// GetSettings wraps a ConfigStore.GetSettings
func (v *Storage) GetSettings() (*Settings, error) {
	return v.src.GetSettings()
}

// SaveSettings wraps a ConfigStore.SaveSettings
func (v *Storage) SaveSettings(s *Settings) error {
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

	return v.src.SaveSettings(s)
}

// GetAuther wraps a ConfigStore.GetAuther
func (v *Storage) GetAuther(t AuthMethod) (Auther, error) {
	auther, err := v.src.GetAuther(t)
	if err != nil {
		return nil, err
	}

	auther.SetStorage(v)
	return auther, nil
}

// SaveAuther wraps a ConfigStore.SaveAuther
func (v *Storage) SaveAuther(a Auther) error {
	return v.src.SaveAuther(a)
}

// GetLinkByHash wraps a Storage.GetLinkByHash.
func (v *Storage) GetLinkByHash(hash string) (*ShareLink, error) {
	return v.src.GetLinkByHash(hash)
}

// GetLinkPermanent wraps a Storage.GetLinkPermanent
func (v *Storage) GetLinkPermanent(path string) (*ShareLink, error) {
	return v.src.GetLinkPermanent(path)
}

// GetLinksByPath wraps a Storage.GetLinksByPath
func (v *Storage) GetLinksByPath(path string) ([]*ShareLink, error) {
	return v.src.GetLinksByPath(path)
}

// SaveLink wraps a Storage.SaveLink
func (v *Storage) SaveLink(s *ShareLink) error {
	return v.src.SaveLink(s)
}

// DeleteLink wraps a Storage.DeleteLink
func (v *Storage) DeleteLink(hash string) error {
	return v.src.DeleteLink(hash)
}

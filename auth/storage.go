package auth

import (
	"github.com/filebrowser/filebrowser/settings"
	"github.com/filebrowser/filebrowser/users"
)

// StorageBackend is a storage backend for auth storage.
type StorageBackend interface {
	Get(settings.AuthMethod) (Auther, error)
	Save(Auther) error
}

// Storage is a auth storage.
type Storage struct {
	back  StorageBackend
	users *users.Storage
}

// NewStorage creates a auth storage from a backend.
func NewStorage(back StorageBackend, users *users.Storage) *Storage {
	return &Storage{back: back, users: users}
}

// Get wraps a StorageBackend.Get and calls SetStorage on the auther.
func (s *Storage) Get(t settings.AuthMethod) (Auther, error) {
	auther, err := s.back.Get(t)
	if err != nil {
		return nil, err
	}

	auther.SetStorage(s.users)
	return auther, nil
}

// Save wraps a StorageBackend.Save.
func (s *Storage) Save(a Auther) error {
	return s.back.Save(a)
}

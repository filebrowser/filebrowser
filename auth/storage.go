package auth

import (
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// StorageBackend is a storage backend for auth storage.
type StorageBackend interface {
	Get(settings.AuthMethod) (Auther, error)
	Save(Auther) error
}

// Storage is a auth storage.
type Storage struct {
	back     StorageBackend
	users    *users.Storage
	settings *settings.Storage
}

// NewStorage creates a auth storage from a backend.
func NewStorage(back StorageBackend, userStore *users.Storage, settingsStore *settings.Storage) *Storage {
	return &Storage{back: back, users: userStore, settings: settingsStore}
}

// Get wraps a StorageBackend.Get.
func (s *Storage) Get(t settings.AuthMethod) (Auther, error) {
	return s.back.Get(t)
}

// Save wraps a StorageBackend.Save.
func (s *Storage) Save(a Auther) error {
	return s.back.Save(a)
}

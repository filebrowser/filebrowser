package bolt

import (
	"github.com/asdine/storm/v3"

	"github.com/rforced/filebrowser/v2/auth"
	"github.com/rforced/filebrowser/v2/settings"
	"github.com/rforced/filebrowser/v2/share"
	"github.com/rforced/filebrowser/v2/storage"
	"github.com/rforced/filebrowser/v2/users"
)

// NewStorage creates a storage.Storage based on Bolt DB.
func NewStorage(db *storm.DB) (*storage.Storage, error) {
	userStore := users.NewStorage(usersBackend{db: db})
	shareStore := share.NewStorage(shareBackend{db: db})
	settingsStore := settings.NewStorage(settingsBackend{db: db})
	authStore := auth.NewStorage(authBackend{db: db}, userStore)

	err := save(db, "version", 2)
	if err != nil {
		return nil, err
	}

	tokenStore := tokenBackend{db: db}

	return &storage.Storage{
		Auth:     authStore,
		Users:    userStore,
		Share:    shareStore,
		Settings: settingsStore,
		Tokens:   tokenStore,
	}, nil
}

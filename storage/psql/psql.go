package psql

import (
	"database/sql"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
)

func ConnectDB(path string) (*sql.DB, error) {
	db, err := sql.Open("postgres", path)
	if err == nil {
		return db, nil
	}
	return nil, err
}

// NewStorage creates a storage.Storage based on Bolt DB.
func NewStorage(db *sql.DB) (*storage.Storage, error) {
	userStore := users.NewStorage(usersBackend{db: db})
	shareStore := share.NewStorage(shareBackend{db: db})
	settingsStore := settings.NewStorage(settingsBackend{db: db})
	authStore := auth.NewStorage(authBackend{db: db}, userStore)

	err := save(db, "version", 2) //nolint:gomnd
	if err != nil {
		return nil, err
	}

	return &storage.Storage{
		Auth:     authStore,
		Users:    userStore,
		Share:    shareStore,
		Settings: settingsStore,
	}, nil
}

func save(db *sql.DB, name string, from interface{}) error {
	// return db.Set("config", name, from)
	return nil
}

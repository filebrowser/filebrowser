package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func IsDBPath(path string) bool {
	prefixes := []string{"sqlite3", "postgres", "mysql"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix+"://") {
			return true
		}
	}
	return false
}

func OpenDB(path string) (*sql.DB, error) {
	prefixes := []string{"sqlite3", "postgres", "mysql"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return connectDB(prefix, strings.TrimPrefix(path, prefix+"://"))
		}
	}
	return nil, errors.New("Unsupported db scheme")
}

func connectDB(dbType string, path string) (*sql.DB, error) {
	db, err := sql.Open(dbType, path)
	if err == nil {
		return db, nil
	}
	return nil, err
}

func NewStorage(db *sql.DB) (*storage.Storage, error) {

	InitUserTable(db)
	InitShareTable(db)
	InitSettingsTable(db)

	userStore := users.NewStorage(usersBackend{db: db})
	shareStore := share.NewStorage(shareBackend{db: db})
	settingsStore := settings.NewStorage(settingsBackend{db: db})
	authStore := auth.NewStorage(authBackend{db: db}, userStore)

	err := SetSetting(db, "version", "2")
	if err != nil {
		fmt.Println("ERROR: fail to set version")
		return nil, err
	}

	// TODO: default
	fmt.Println(GetSetting(db, "auther"))
	if GetSetting(db, "auther") == "" {
		err := SetSetting(db, "auther", "json")
		if err != nil {
			fmt.Println("ERROR: fail to set auther")
			return nil, err
		}
	}

	storage := &storage.Storage{
		Auth:     authStore,
		Users:    userStore,
		Share:    shareStore,
		Settings: settingsStore,
	}
	return storage, nil
}

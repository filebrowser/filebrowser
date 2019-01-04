package storage

import (
	"github.com/filebrowser/filebrowser/auth"
	"github.com/filebrowser/filebrowser/settings"
	"github.com/filebrowser/filebrowser/share"
	"github.com/filebrowser/filebrowser/users"
)

// Storage is a storage powered by a Backend whih makes the neccessary
// verifications when fetching and saving data to ensure consistency.
type Storage struct {
	Users *users.Storage
	Share *share.Storage
	Auth *auth.Storage
	Settings *settings.Storage
}

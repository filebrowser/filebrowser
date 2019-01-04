package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/settings"
	"github.com/filebrowser/filebrowser/users"
)

// MethodNoAuth is used to identify no auth.
const MethodNoAuth settings.AuthMethod = "noauth"

// NoAuth is no auth implementation of auther.
type NoAuth struct {
	storage *users.Storage
}

// Auth uses authenticates user 1.
func (a *NoAuth) Auth(r *http.Request) (*users.User, error) {
	return a.storage.Get(1)
}

// SetStorage attaches the storage to the auther.
func (a *NoAuth) SetStorage(s *users.Storage) {
	a.storage = s
}

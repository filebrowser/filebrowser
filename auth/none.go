package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/types"
)

// MethodNoAuth is used to identify no auth.
const MethodNoAuth types.AuthMethod = "noauth"

// NoAuth is no auth implementation of auther.
type NoAuth struct {
	store *types.Storage
}

// Auth uses authenticates user 1.
func (a *NoAuth) Auth(r *http.Request) (*types.User, error) {
	return a.store.GetUser(1)
}

// SetStorage attaches the storage information to the auther.
func (a *NoAuth) SetStorage(s *types.Storage) {
	a.store = s
}

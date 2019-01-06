package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// MethodNoAuth is used to identify no auth.
const MethodNoAuth settings.AuthMethod = "noauth"

// NoAuth is no auth implementation of auther.
type NoAuth struct {
}

// Auth uses authenticates user 1.
func (a *NoAuth) Auth(r *http.Request, sto *users.Storage, set *settings.Settings) (*users.User, error) {
	return sto.Get(set.Scope, 1)
}

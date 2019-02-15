package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// MethodNoAuth is used to identify no auth.
const MethodNoAuth settings.AuthMethod = "noauth"

// NoAuth is no auth implementation of auther.
type NoAuth struct{}

// Auth uses authenticates user 1.
func (a NoAuth) Auth(r *http.Request, sto *users.Storage, root string) (*users.User, error) {
	return sto.Get(root, uint(1))
}

// LoginPage tells that no auth doesn't require a login page.
func (a NoAuth) LoginPage() bool {
	return false
}

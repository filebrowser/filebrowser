package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/types"
)

// MethodNoAuth is used to identify no auth.
const MethodNoAuth types.AuthMethod = "noauth"

// NoAuth is no auth implementation of auther.
type NoAuth struct {
	Store *types.UsersVerify `json:"-"`
}

// Auth uses authenticates user 1.
func (a NoAuth) Auth(r *http.Request) (*types.User, error) {
	return a.Store.Get(1)
}

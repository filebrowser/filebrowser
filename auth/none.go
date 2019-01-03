package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/types"
)

// MethodNoAuth is used to identify no auth.
const MethodNoAuth types.AuthMethod = "noauth"

// NoAuth is no auth implementation of auther.
type NoAuth struct {
	instance *types.FileBrowser
}

// Auth uses authenticates user 1.
func (a *NoAuth) Auth(r *http.Request) (*types.User, error) {
	return a.instance.GetUser(1)
}

// SetInstance attaches the instance to the auther.
func (a *NoAuth) SetInstance(i *types.FileBrowser) {
	a.instance = i
}

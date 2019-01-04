package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/lib"
)

// MethodNoAuth is used to identify no auth.
const MethodNoAuth lib.AuthMethod = "noauth"

// NoAuth is no auth implementation of auther.
type NoAuth struct {
	instance *lib.FileBrowser
}

// Auth uses authenticates user 1.
func (a *NoAuth) Auth(r *http.Request) (*lib.User, error) {
	return a.instance.GetUser(1)
}

// SetInstance attaches the instance to the auther.
func (a *NoAuth) SetInstance(i *lib.FileBrowser) {
	a.instance = i
}

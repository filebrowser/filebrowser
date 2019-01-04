package auth

import (
	"net/http"
	"os"

	"github.com/filebrowser/filebrowser/settings"
	"github.com/filebrowser/filebrowser/users"
	"github.com/filebrowser/filebrowser/errors"
)

// MethodProxyAuth is used to identify no auth.
const MethodProxyAuth settings.AuthMethod = "proxy"

// ProxyAuth is a proxy implementation of an auther.
type ProxyAuth struct {
	Header  string
	storage *users.Storage
}

// Auth authenticates the user via an HTTP header.
func (a *ProxyAuth) Auth(r *http.Request) (*users.User, error) {
	username := r.Header.Get(a.Header)
	user, err := a.storage.Get(username)
	if err == errors.ErrNotExist {
		return nil, os.ErrPermission
	}

	return user, err
}

// SetStorage attaches the storage to the auther.
func (a *ProxyAuth) SetStorage(s *users.Storage) {
	a.storage = s
}

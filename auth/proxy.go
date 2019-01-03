package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/types"
)

// MethodProxyAuth is used to identify no auth.
const MethodProxyAuth types.AuthMethod = "proxy"

// ProxyAuth is a proxy implementation of an auther.
type ProxyAuth struct {
	Header string
	store  *types.Storage
}

// Auth authenticates the user via an HTTP header.
func (a *ProxyAuth) Auth(r *http.Request) (*types.User, error) {
	username := r.Header.Get(a.Header)
	user, err := a.store.GetUser(username)
	if err == types.ErrNotExist {
		return nil, types.ErrNoPermission
	}

	return user, err
}

// SetStorage attaches the storage information to the auther.
func (a *ProxyAuth) SetStorage(s *types.Storage) {
	a.store = s
}

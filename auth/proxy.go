package auth

import (
	"net/http"
	"os"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// MethodProxyAuth is used to identify no auth.
const MethodProxyAuth settings.AuthMethod = "proxy"

// ProxyAuth is a proxy implementation of an auther.
type ProxyAuth struct {
	Header string
}

// Auth authenticates the user via an HTTP header.
func (a *ProxyAuth) Auth(r *http.Request, sto *users.Storage, set *settings.Settings) (*users.User, error) {
	username := r.Header.Get(a.Header)
	user, err := sto.Get(set.Scope, username)
	if err == errors.ErrNotExist {
		return nil, os.ErrPermission
	}

	return user, err
}

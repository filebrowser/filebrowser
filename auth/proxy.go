package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/types"
)

// MethodProxyAuth is used to identify no auth.
const MethodProxyAuth types.AuthMethod = "proxy"

// ProxyAuth is a proxy implementation of an auther.
type ProxyAuth struct {
	Header   string
	instance *types.FileBrowser
}

// Auth authenticates the user via an HTTP header.
func (a *ProxyAuth) Auth(r *http.Request) (*types.User, error) {
	username := r.Header.Get(a.Header)
	user, err := a.instance.GetUser(username)
	if err == types.ErrNotExist {
		return nil, types.ErrNoPermission
	}

	return user, err
}

// SetInstance attaches the instance to the auther.
func (a *ProxyAuth) SetInstance(i *types.FileBrowser) {
	a.instance = i
}

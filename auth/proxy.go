package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/lib"
)

// MethodProxyAuth is used to identify no auth.
const MethodProxyAuth lib.AuthMethod = "proxy"

// ProxyAuth is a proxy implementation of an auther.
type ProxyAuth struct {
	Header   string
	instance *lib.FileBrowser
}

// Auth authenticates the user via an HTTP header.
func (a *ProxyAuth) Auth(r *http.Request) (*lib.User, error) {
	username := r.Header.Get(a.Header)
	user, err := a.instance.GetUser(username)
	if err == lib.ErrNotExist {
		return nil, lib.ErrNoPermission
	}

	return user, err
}

// SetInstance attaches the instance to the auther.
func (a *ProxyAuth) SetInstance(i *lib.FileBrowser) {
	a.instance = i
}

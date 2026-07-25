package auth

import (
	"errors"
	"net/http"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// MethodProxyAuth is used to identify no auth.
const MethodProxyAuth settings.AuthMethod = "proxy"

// ProxyAuth is a proxy implementation of an auther.
type ProxyAuth struct {
	Header string `json:"header"`
}

// Auth authenticates the user via an HTTP header.
func (a ProxyAuth) Auth(r *http.Request, usr users.Store, setting *settings.Settings, srv *settings.Server) (*users.User, error) {
	username := r.Header.Get(a.Header)
	user, err := usr.Get(srv.Root, srv.FollowExternalSymlinks, username)
	if errors.Is(err, fberrors.ErrNotExist) {
		return a.createUser(usr, setting, srv, username)
	}
	return user, err
}

func (a ProxyAuth) createUser(usr users.Store, setting *settings.Settings, srv *settings.Server, username string) (*users.User, error) {
	const randomPasswordLength = settings.DefaultMinimumPasswordLength + 10
	pwd, err := users.RandomPwd(randomPasswordLength)
	if err != nil {
		return nil, err
	}

	var hashedRandomPassword string
	hashedRandomPassword, err = users.ValidateAndHashPwd(pwd, setting.MinimumPasswordLength)
	if err != nil {
		return nil, err
	}

	user := &users.User{
		Username:     username,
		Password:     hashedRandomPassword,
		LockPassword: true,
	}
	setting.Defaults.Apply(user)
	user.Perm.Admin = false
	user.Perm.Execute = false
	user.Commands = []string{}

	if err = setting.CreateUserHome(user, usr, srv.Root, false); err != nil {
		return nil, err
	}

	if err = usr.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}

// LoginPage tells that proxy auth doesn't require a login page.
func (a ProxyAuth) LoginPage() bool {
	return false
}

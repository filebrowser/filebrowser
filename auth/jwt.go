package auth

import (
	"context"
	nerrors "errors"
	"net/http"
	"os"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// MethodJWTAuth is used to identify JWTAuth auth.
const MethodJWTAuth settings.AuthMethod = "jwt-header"

// JWTAuth is a JWTAuth implementation of an auther.
type JWTAuth struct {
	CertsURL      string `json:"certsurl"`
	Aud           string `json:"aud"`
	Iss           string `json:"iss"`
	UsernameClaim string `json:"usernameClaim"`
	Header        string `json:"header"`
	remoteKeySet  *oidc.RemoteKeySet
	init          sync.Once
}

// Auth authenticates the user via a JWT token in an HTTP header.
func (a *JWTAuth) Auth(r *http.Request, usr users.Store, stg *settings.Settings, srv *settings.Server) (*users.User, error) {
	a.init.Do(func() {
		a.remoteKeySet = oidc.NewRemoteKeySet(context.Background(), a.CertsURL)
	})

	accessJWT := r.Header.Get(a.Header)
	if accessJWT == "" {
		return nil, os.ErrPermission
	}

	// The Application Audience (AUD) tag for your application
	config := &oidc.Config{
		ClientID: a.Aud,
	}

	verifier := oidc.NewVerifier(a.Iss, a.remoteKeySet, config)

	token, err := verifier.Verify(r.Context(), accessJWT)
	if err != nil {
		return nil, os.ErrPermission
	}

	payload := map[string]any{}
	err = token.Claims(&payload)
	if err != nil {
		return nil, os.ErrPermission
	}

	user, err := usr.Get(srv.Root, payload[a.UsernameClaim])
	if nerrors.Is(err, errors.ErrNotExist) {
		return nil, os.ErrPermission
	}

	return user, err
}

// LoginPage tells that proxy auth doesn't require a login page.
func (a *JWTAuth) LoginPage() bool {
	return false
}

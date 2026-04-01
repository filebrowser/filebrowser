package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// MethodOIDCAuth is used to identify OIDC authentication.
const MethodOIDCAuth settings.AuthMethod = "oidc"

// OIDCAuth implements OIDC/OAuth2 authentication via an external provider (e.g. Authelia).
// Config values can be overridden by environment variables:
//
//	FB_OIDC_ISSUER, FB_OIDC_CLIENT_ID, FB_OIDC_CLIENT_SECRET, FB_OIDC_USERNAME_CLAIM
type OIDCAuth struct {
	IssuerURL     string `json:"issuerUrl"`
	ClientID      string `json:"clientId"`
	ClientSecret  string `json:"clientSecret"`
	UsernameClaim string `json:"usernameClaim"`
}

// Auth is not used for OIDC – the redirect flow handles authentication.
// Authentication happens via /api/auth/oidc and /api/auth/oidc/callback.
func (a *OIDCAuth) Auth(_ *http.Request, _ users.Store, _ *settings.Settings, _ *settings.Server) (*users.User, error) {
	return nil, os.ErrPermission
}

// LoginPage returns true because OIDC shows a login page with a SSO button.
func (a *OIDCAuth) LoginPage() bool {
	return true
}

func (a *OIDCAuth) getIssuerURL() string {
	if v := os.Getenv("FB_OIDC_ISSUER"); v != "" {
		return v
	}
	return a.IssuerURL
}

func (a *OIDCAuth) getClientID() string {
	if v := os.Getenv("FB_OIDC_CLIENT_ID"); v != "" {
		return v
	}
	return a.ClientID
}

func (a *OIDCAuth) getClientSecret() string {
	if v := os.Getenv("FB_OIDC_CLIENT_SECRET"); v != "" {
		return v
	}
	return a.ClientSecret
}

func (a *OIDCAuth) getUsernameClaim() string {
	if v := os.Getenv("FB_OIDC_USERNAME_CLAIM"); v != "" {
		return v
	}
	if a.UsernameClaim != "" {
		return a.UsernameClaim
	}
	return "preferred_username"
}

// GetProvider fetches the OIDC provider via discovery.
func (a *OIDCAuth) GetProvider(ctx context.Context) (*oidc.Provider, error) {
	issuer := a.getIssuerURL()
	if issuer == "" {
		return nil, errors.New("OIDC issuer URL not configured (set FB_OIDC_ISSUER or auth.oidc.issuer)")
	}
	return oidc.NewProvider(ctx, issuer)
}

// GetOAuthConfig returns the oauth2.Config for the OIDC flow.
func (a *OIDCAuth) GetOAuthConfig(provider *oidc.Provider, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     a.getClientID(),
		ClientSecret: a.getClientSecret(),
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
}

// HandleCallback exchanges the OIDC code for a user, creating the user if needed.
func (a *OIDCAuth) HandleCallback(
	ctx context.Context,
	code string,
	redirectURL string,
	usr users.Store,
	setting *settings.Settings,
	srv *settings.Server,
) (*users.User, error) {
	provider, err := a.GetProvider(ctx)
	if err != nil {
		return nil, err
	}

	oauth2Config := a.GetOAuthConfig(provider, redirectURL)

	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: a.getClientID()})
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token in token response")
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	// Merge UserInfo claims — Authelia (and some other providers) only include
	// preferred_username in the UserInfo endpoint, not in the ID token itself.
	userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err == nil {
		var userInfoClaims map[string]interface{}
		if userInfo.Claims(&userInfoClaims) == nil {
			for k, v := range userInfoClaims {
				if _, exists := claims[k]; !exists {
					claims[k] = v
				}
			}
		}
	}

	username := a.extractUsername(claims)
	if username == "" {
		return nil, fmt.Errorf("could not determine username from OIDC claims (tried claim %q)", a.getUsernameClaim())
	}

	user, err := usr.Get(srv.Root, username)
	if errors.Is(err, fberrors.ErrNotExist) {
		return createOIDCUser(usr, setting, srv, username)
	}
	return user, err
}

func (a *OIDCAuth) extractUsername(claims map[string]interface{}) string {
	claim := a.getUsernameClaim()
	if val, ok := claims[claim]; ok {
		if s, ok := val.(string); ok && s != "" {
			return s
		}
	}
	// Fallback to email
	if email, ok := claims["email"]; ok {
		if s, ok := email.(string); ok {
			return s
		}
	}
	return ""
}

func createOIDCUser(usr users.Store, setting *settings.Settings, srv *settings.Server, username string) (*users.User, error) {
	const randomPasswordLength = settings.DefaultMinimumPasswordLength + 10
	pwd, err := users.RandomPwd(randomPasswordLength)
	if err != nil {
		return nil, err
	}

	hashedPwd, err := users.ValidateAndHashPwd(pwd, setting.MinimumPasswordLength)
	if err != nil {
		return nil, err
	}

	user := &users.User{
		Username:     username,
		Password:     hashedPwd,
		LockPassword: true,
	}
	setting.Defaults.Apply(user)

	userHome, err := setting.MakeUserDir(user.Username, user.Scope, srv.Root)
	if err != nil {
		return nil, err
	}
	user.Scope = userHome

	if err := usr.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

// GenerateState generates a random state value for CSRF protection.
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

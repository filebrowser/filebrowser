package auth

import (
	"context"
	"fmt"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// MethodOIDCAuth is used to identify oidc auth.
const MethodOIDCAuth settings.AuthMethod = "oidc"

// OIDCAuth is an Open ID Connect auther implementation.
type OIDCAuth struct {
	OIDC *OAuthClient `json:"oidc" yaml:"oidc"`
}

// Auth is executed when the identity provider enters the callback phase of an oauth code flow.
func (a OIDCAuth) Auth(r *http.Request, usr users.Store, _ *settings.Settings, srv *settings.Server) (*users.User, error) {
	cookie, _ := r.Cookie("auth")
	if cookie != nil {
		return nil, nil
	}

	log.Println("oidc auth callback")
	u, err := a.OIDC.HandleAuthCallback(r, usr, srv)

	return u, err
}

// LoginPage tells that oidc auth doesn't require a login page.
func (a OIDCAuth) LoginPage() bool {
	return false
}

// OAuthClient describes the oidc connector parameters.
type OAuthClient struct {
	ClientID     string                `json:"clientID"`
	ClientSecret string                `json:"clientSecret"`
	Issuer       string                `json:"issuer"`
	RedirectURL  string                `json:"redirectURL"`
	OAuth2Config oauth2.Config         `json:"-"`
	Verifier     *oidc.IDTokenVerifier `json:"-"`
}

// InitClient configures the connector via oidc discovery.
func (o *OAuthClient) InitClient() {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, o.Issuer)
	if err != nil {
		log.Fatal(err)
	}

	o.Verifier = provider.Verifier(&oidc.Config{ClientID: o.ClientID})
	o.OAuth2Config = oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		RedirectURL:  o.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
}

// InitAuthFlow triggers the oidc authentication flow.
func (o *OAuthClient) InitAuthFlow(w http.ResponseWriter, r *http.Request) {
	o.InitClient()
	state := fmt.Sprintf("%x", rand.Uint32())
	nonce := fmt.Sprintf("%x", rand.Uint32())
	o.OAuth2Config.RedirectURL += "?redirect=" + r.URL.Path
	url := o.OAuth2Config.AuthCodeURL(state, oidc.Nonce(nonce))

	log.Println("oidc init flow ", url)
	w.Header().Set("Set-Cookie", "state="+state+"; path=/")
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

// HandleAuthCallback manages code exchange and obtains the id token.
func (o *OAuthClient) HandleAuthCallback(r *http.Request, usr users.Store, srv *settings.Server) (*users.User, error) {
	o.InitClient()
	code := r.URL.Query().Get("code")
	stateQuery := r.URL.Query().Get("state")
	stateCookie, err := r.Cookie("state")

	// Validate state
	if code == "" || stateQuery == "" || err != nil || stateQuery != stateCookie.Value {
		log.Fatal("Invalid request")
		return nil, os.ErrPermission
	}

	// Exchange code for token
	oauth2Token, err := o.OAuth2Config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Parse id token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Fatal("Invalid token")
		return nil, os.ErrPermission
	}

	// Verify id token
	idToken, err := o.Verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		log.Fatal("oidc verify failed")
		return nil, err
	}

	// Extract claims
	var claims struct {
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
		Username string `json:"preferred_username"`
		Profile  string `json:"profile"`
	}
	if err := idToken.Claims(&claims); err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Find filebrowser user by oidc username
	u, err := usr.Get(srv.Root, claims.Username)
	if err != nil {
		log.Println("oidc authenticated but no matching filebrowser user")
		return nil, os.ErrPermission
	}
	u.AuthSource = "oidc"
	log.Println("oidc success (user, claims) ", u.Username, claims)

	return u, nil
}

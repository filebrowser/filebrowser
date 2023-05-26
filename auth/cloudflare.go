package auth

import (
	"fmt"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// MethodCloudflareAuth is used to identify no auth.
const MethodCloudflareAuth settings.AuthMethod = "cloudflare-access"

// CloudflareAuth is a proxy implementation of an auther.
type CloudflareAuth struct {
	Team string `json:"team"`
	Aud  string `json:"aud"`
}

type CloudflareTokenPayload struct {
	Email string
}

// Auth authenticates the user via an HTTP header.
func (a CloudflareAuth) Auth(r *http.Request, usr users.Store, stg *settings.Settings, srv *settings.Server) (*users.User, error) {
	accessJWT := r.Header.Get("Cf-Access-Jwt-Assertion")
	if accessJWT == "" {
		return nil, os.ErrPermission
	}

	// The Application Audience (AUD) tag for your application
	config := &oidc.Config{
		ClientID: a.Aud,
	}

	teamDomain := fmt.Sprintf("https://%s.cloudflareaccess.com", a.Team)
	certsURL := fmt.Sprintf("%s/cdn-cgi/access/certs", teamDomain)
	keySet := oidc.NewRemoteKeySet(r.Context(), certsURL)
	verifier := oidc.NewVerifier(teamDomain, keySet, config)

	token, err := verifier.Verify(r.Context(), accessJWT)
	if err != nil {
		return nil, os.ErrPermission
	}

	payload := new(CloudflareTokenPayload)
	token.Claims(&payload)

	user, err := usr.Get(srv.Root, payload.Email)
	if err == errors.ErrNotExist {
		return nil, os.ErrPermission
	}

	return user, err
}

// LoginPage tells that proxy auth doesn't require a login page.
func (a CloudflareAuth) LoginPage() bool {
	return false
}

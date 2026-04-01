package fbhttp

import (
	"fmt"
	"net/http"
	"time"

	fbAuth "github.com/filebrowser/filebrowser/v2/auth"
)

const oidcStateCookie = "oidc_state"

// oidcRedirectHandler initiates the OIDC login flow by redirecting to the provider.
var oidcRedirectHandler handleFunc = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	auther, err := d.store.Auth.Get(d.settings.AuthMethod)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	oidcAuth, ok := auther.(*fbAuth.OIDCAuth)
	if !ok {
		return http.StatusBadRequest, fmt.Errorf("auth method is not oidc")
	}

	state, err := fbAuth.GenerateState()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Persist state in a short-lived cookie for CSRF validation
	http.SetCookie(w, &http.Cookie{
		Name:     oidcStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	provider, err := oidcAuth.GetProvider(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	redirectURL := oidcCallbackURL(r, d)
	oauth2Config := oidcAuth.GetOAuthConfig(provider, redirectURL)
	authURL := oauth2Config.AuthCodeURL(state)

	http.Redirect(w, r, authURL, http.StatusFound)
	return 0, nil
}

// oidcCallbackHandler processes the OIDC callback: exchanges code, issues JWT,
// then redirects the browser to the frontend login page with the token.
func oidcCallbackHandler(tokenExpireTime time.Duration) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		// Validate CSRF state
		stateCookie, err := r.Cookie(oidcStateCookie)
		if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
			return http.StatusForbidden, fmt.Errorf("invalid OIDC state")
		}

		// Clear state cookie
		http.SetCookie(w, &http.Cookie{
			Name:   oidcStateCookie,
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})

		auther, err := d.store.Auth.Get(d.settings.AuthMethod)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		oidcAuth, ok := auther.(*fbAuth.OIDCAuth)
		if !ok {
			return http.StatusBadRequest, fmt.Errorf("auth method is not oidc")
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			return http.StatusBadRequest, fmt.Errorf("missing OIDC code")
		}

		user, err := oidcAuth.HandleCallback(
			r.Context(),
			code,
			oidcCallbackURL(r, d),
			d.store.Users,
			d.settings,
			d.server,
		)
		if err != nil {
			return http.StatusForbidden, err
		}

		// Generate JWT token (reuse printToken logic, but capture output)
		jwt, err := createToken(d, user, tokenExpireTime)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		// Redirect to the login page with the token so the frontend can store it
		redirectTarget := d.server.BaseURL + "/login?token=" + jwt
		http.Redirect(w, r, redirectTarget, http.StatusFound)
		return 0, nil
	}
}

// oidcCallbackURL builds the absolute callback URL registered with the OIDC provider.
func oidcCallbackURL(r *http.Request, d *data) string {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	host := r.Host
	if fwdHost := r.Header.Get("X-Forwarded-Host"); fwdHost != "" {
		host = fwdHost
	}
	return fmt.Sprintf("%s://%s%s/api/auth/oidc/callback", scheme, host, d.server.BaseURL)
}

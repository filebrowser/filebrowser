package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang-jwt/jwt/v4/request"
)

type totpUserInfo struct {
	ID uint `json:"id"`
}

type totpAuthToken struct {
	User totpUserInfo `json:"user"`
	jwt.RegisteredClaims
}

type totpExtractor []string

func (e totpExtractor) ExtractToken(r *http.Request) (string, error) {
	token, _ := request.HeaderExtractor{"X-TOTP-Auth"}.ExtractToken(r)

	// Checks if the token isn't empty and if it contains two dots.
	// The former prevents incompatibility with URLs that previously
	// used basic auth.
	if token != "" && strings.Count(token, ".") == 2 {
		return token, nil
	}

	return "", request.ErrNoTokenInRequest
}

func verifyTOTPHandler(tokenExpireTime time.Duration) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		code := r.Header.Get("X-TOTP-CODE")
		if code == "" {
			return http.StatusUnauthorized, nil
		}

		keyFunc := func(_ *jwt.Token) (interface{}, error) {
			return d.settings.Key, nil
		}

		var tk totpAuthToken
		token, err := request.ParseFromRequest(r, &totpExtractor{}, keyFunc, request.WithClaims(&tk))

		if err != nil || !token.Valid {
			return http.StatusUnauthorized, nil
		}

		d.user, err = d.store.Users.Get(d.server.Root, tk.User.ID)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if ok, err := users.CheckTOTP(d.server.TOTPEncryptionKey, d.user.TOTPSecret, d.user.TOTPNonce, code); err != nil {
			return http.StatusInternalServerError, err
		} else if !ok {
			return http.StatusUnauthorized, nil
		}

		return printToken(w, r, d, d.user, tokenExpireTime)
	}
}

func withTOTP(fn handleFunc) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if d.user.TOTPSecret == "" {
			return fn(w, r, d)
		}

		if code := r.Header.Get("X-TOTP-CODE"); code == "" {
			return http.StatusForbidden, nil
		} else {
			if ok, err := users.CheckTOTP(d.server.TOTPEncryptionKey, d.user.TOTPSecret, d.user.TOTPNonce, code); err != nil {
				return http.StatusInternalServerError, err
			} else if !ok {
				return http.StatusForbidden, nil
			}

			return fn(w, r, d)
		}
	})
}

func printTOTPToken(w http.ResponseWriter, _ *http.Request, d *data, user *users.User, tokenExpirationTime time.Duration) (int, error) {
	claims := &totpAuthToken{
		User: totpUserInfo{
			ID: user.ID,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpirationTime)),
			Issuer:    "File Browser TOTP",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(d.settings.Key)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, nil, loginResponse{Token: signed, OTP: true})
}

package http

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	fm "github.com/hacdias/filemanager"
)

// authHandler proccesses the authentication for the user.
func authHandler(c *fm.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	// NoAuth instances shouldn't call this method.
	if c.NoAuth {
		return 0, nil
	}

	// Receive the credentials from the request and unmarshal them.
	var cred fm.User
	if r.Body == nil {
		return http.StatusForbidden, nil
	}

	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		return http.StatusForbidden, nil
	}

	// Checks if the user exists.
	u, ok := c.Users[cred.Username]
	if !ok {
		return http.StatusForbidden, nil
	}

	// Checks if the password is correct.
	if !checkPasswordHash(cred.Password, u.Password) {
		return http.StatusForbidden, nil
	}

	c.User = u
	return printToken(c, w)
}

// renewAuthHandler is used when the front-end already has a JWT token
// and is checking if it is up to date. If so, updates its info.
func renewAuthHandler(c *fm.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	ok, u := validateAuth(c, r)
	if !ok {
		return http.StatusForbidden, nil
	}

	c.User = u
	return printToken(c, w)
}

// claims is the JWT claims.
type claims struct {
	fm.User
	NoAuth bool `json:"noAuth"`
	jwt.StandardClaims
}

// printToken prints the final JWT token to the user.
func printToken(c *fm.Context, w http.ResponseWriter) (int, error) {
	// Creates a copy of the user and removes it password
	// hash so it never arrives to the user.
	u := fm.User{}
	u = *c.User
	u.Password = ""

	// Builds the claims.
	claims := claims{
		u,
		c.NoAuth,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    "File Manager",
		},
	}

	// Creates the token and signs it.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(c.key)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Writes the token.
	w.Header().Set("Content-Type", "cty")
	w.Write([]byte(signed))
	return 0, nil
}

type extractor []string

func (e extractor) ExtractToken(r *http.Request) (string, error) {
	token, _ := request.AuthorizationHeaderExtractor.ExtractToken(r)

	// Checks if the token isn't empty and if it contains two dots.
	// The former prevents incompatibility with URLs that previously
	// used basic auth.
	if token != "" && strings.Count(token, ".") == 2 {
		return token, nil
	}

	cookie, err := r.Cookie("auth")
	if err != nil {
		return "", request.ErrNoTokenInRequest
	}

	return cookie.Value, nil
}

// validateAuth is used to validate the authentication and returns the
// User if it is valid.
func validateAuth(c *fm.Context, r *http.Request) (bool, *fm.User) {
	if c.NoAuth {
		c.User = c.DefaultUser
		return true, c.User
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return c.key, nil
	}
	var claims claims
	token, err := request.ParseFromRequestWithClaims(r,
		extractor{},
		&claims,
		keyFunc,
	)

	if err != nil || !token.Valid {
		return false, nil
	}

	u, ok := c.Users[claims.User.Username]
	if !ok {
		return false, nil
	}

	c.User = u
	return true, u
}

// hashPassword generates an hash from a password using bcrypt.
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPasswordHash compares a password with an hash to check if they match.
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

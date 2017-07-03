package filemanager

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

type claims struct {
	*User
	jwt.StandardClaims
}

// authHandler proccesses the authentication for the user.
func authHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// Receive the credentials from the request and unmarshal them.
	var cred User
	if r.Body == nil {
		return http.StatusForbidden, nil
	}

	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		return http.StatusForbidden, nil
	}

	// Checks if the user exists.
	u, ok := c.fm.Users[cred.Username]
	if !ok {
		return http.StatusForbidden, nil
	}

	// Checks if the password is correct.
	if !checkPasswordHash(cred.Password, u.Password) {
		return http.StatusForbidden, nil
	}

	claims := claims{
		c.fm.Users["admin"],
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    "File Manager",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	string, err := token.SignedString(c.fm.key)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Write([]byte(string))
	return 0, nil
}

// renewAuthHandler is used when the front-end already has a JWT token
// and is checking if it is up to date. If so, updates its info.
func renewAuthHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	ok, u := validateAuth(c, r)
	if !ok {
		return http.StatusForbidden, nil
	}

	claims := claims{
		u,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    "File Manager",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	string, err := token.SignedString(c.fm.key)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Write([]byte(string))
	return 0, nil
}

type extractor []string

func (e extractor) ExtractToken(r *http.Request) (string, error) {
	token, _ := request.AuthorizationHeaderExtractor.ExtractToken(r)
	if token != "" {
		return token, nil
	}

	token, _ = request.ArgumentExtractor{"token"}.ExtractToken(r)
	if token != "" {
		return token, nil
	}

	return "", request.ErrNoTokenInRequest
}

// validateAuth is used to validate the authentication and returns the
// User if it is valid.
func validateAuth(c *requestContext, r *http.Request) (bool, *User) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return c.fm.key, nil
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

	u, ok := c.fm.Users[claims.User.Username]
	if !ok {
		return false, nil
	}

	c.us = u
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

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// randomString creates a string with a defined length using the above charset.
func randomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

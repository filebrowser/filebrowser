package share

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
)

var shareHashRegex = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type CreateBody struct {
	Hash     string `json:"hash"`
	Password string `json:"password"`
	Expires  string `json:"expires"`
	Unit     string `json:"unit"`
}

// Link is the information needed to build a shareable link.
type Link struct {
	Hash         string `json:"hash" storm:"id,index"`
	Path         string `json:"path" storm:"index"`
	UserID       uint   `json:"userID"`
	Expire       int64  `json:"expire"`
	PasswordHash string `json:"password_hash,omitempty"`
	// Token is a random value that will only be set when PasswordHash is set. It is
	// URL-Safe and is used to download links in password-protected shares via a
	// query arg.
	Token string `json:"token,omitempty"`
}

func GenerateHash() (string, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func ValidateHash(hash string) (string, error) {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return "", fmt.Errorf("share key cannot be empty: %w", fberrors.ErrInvalidRequestParams)
	}

	if !shareHashRegex.MatchString(hash) {
		return "", fmt.Errorf(
			"share key must use only letters, numbers, hyphen, or underscore: %w",
			fberrors.ErrInvalidRequestParams,
		)
	}

	return hash, nil
}

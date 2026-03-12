package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Token struct {
	Token     string    `json:"token" storm:"id"`
	UserID    uint      `json:"userID" storm:"index"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

type TokenStore interface {
	Save(t *Token) error
	Get(token string) (*Token, error)
	Delete(token string) error
	DeleteByUser(userID uint) error
	DeleteExpired() error
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

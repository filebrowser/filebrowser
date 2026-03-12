package bolt

import (
	"errors"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"

	"github.com/filebrowser/filebrowser/v2/auth"
	fberrors "github.com/filebrowser/filebrowser/v2/errors"
)

type tokenBackend struct {
	db *storm.DB
}

func (s tokenBackend) Save(t *auth.Token) error {
	return s.db.Save(t)
}

func (s tokenBackend) Get(token string) (*auth.Token, error) {
	var t auth.Token
	err := s.db.One("Token", token, &t)
	if errors.Is(err, storm.ErrNotFound) {
		return nil, fberrors.ErrNotExist
	}
	return &t, err
}

func (s tokenBackend) Delete(token string) error {
	err := s.db.DeleteStruct(&auth.Token{Token: token})
	if errors.Is(err, storm.ErrNotFound) {
		return nil
	}
	return err
}

func (s tokenBackend) DeleteByUser(userID uint) error {
	var tokens []auth.Token
	err := s.db.Select(q.Eq("UserID", userID)).Find(&tokens)
	if errors.Is(err, storm.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, t := range tokens {
		if err := s.db.DeleteStruct(&t); err != nil {
			return err
		}
	}
	return nil
}

func (s tokenBackend) DeleteExpired() error {
	var tokens []auth.Token
	err := s.db.Select(q.Lt("ExpiresAt", time.Now())).Find(&tokens)
	if errors.Is(err, storm.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, t := range tokens {
		if err := s.db.DeleteStruct(&t); err != nil {
			return err
		}
	}
	return nil
}

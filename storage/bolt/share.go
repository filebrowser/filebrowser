package bolt

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/share"
)

type shareBackend struct {
	db *storm.DB
}

func (s shareBackend) GetByHash(hash string) (*share.Link, error) {
	var v share.Link
	err := s.db.One("Hash", hash, &v)
	if err == storm.ErrNotFound {
		return nil, errors.ErrNotExist
	}

	return &v, err
}

func (s shareBackend) GetPermanent(path string, id uint) (*share.Link, error) {
	var v share.Link
	err := s.db.Select(q.Eq("Path", path), q.Eq("Expire", 0), q.Eq("UserID", id)).First(&v)
	if err == storm.ErrNotFound {
		return nil, errors.ErrNotExist
	}

	return &v, err
}

func (s shareBackend) Gets(path string, id uint) ([]*share.Link, error) {
	var v []*share.Link
	err := s.db.Select(q.Eq("Path", path), q.Eq("UserID", id)).Find(&v)
	if err == storm.ErrNotFound {
		return v, errors.ErrNotExist
	}

	return v, err
}

func (s shareBackend) Save(l *share.Link) error {
	return s.db.Save(l)
}

func (s shareBackend) Delete(hash string) error {
	err := s.db.DeleteStruct(&share.Link{Hash: hash})
	if err == storm.ErrNotFound {
		return nil
	}
	return err
}

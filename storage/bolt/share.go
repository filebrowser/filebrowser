package bolt

import (
	"errors"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	bolt "go.etcd.io/bbolt"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/share"
)

type shareBackend struct {
	db *storm.DB
}

func (s shareBackend) All() ([]*share.Link, error) {
	var v []*share.Link
	err := s.db.All(&v)
	if errors.Is(err, storm.ErrNotFound) {
		return v, fberrors.ErrNotExist
	}

	return v, err
}

func (s shareBackend) FindByUserID(id uint) ([]*share.Link, error) {
	var v []*share.Link
	err := s.db.Select(q.Eq("UserID", id)).Find(&v)
	if errors.Is(err, storm.ErrNotFound) {
		return v, fberrors.ErrNotExist
	}

	return v, err
}

func (s shareBackend) GetByHash(hash string) (*share.Link, error) {
	var v share.Link
	err := s.db.One("Hash", hash, &v)
	if errors.Is(err, storm.ErrNotFound) {
		return nil, fberrors.ErrNotExist
	}

	return &v, err
}

func (s shareBackend) GetPermanent(path string, id uint) (*share.Link, error) {
	var v share.Link
	err := s.db.Select(q.Eq("Path", path), q.Eq("Expire", 0), q.Eq("UserID", id)).First(&v)
	if errors.Is(err, storm.ErrNotFound) {
		return nil, fberrors.ErrNotExist
	}

	return &v, err
}

func (s shareBackend) Gets(path string, id uint) ([]*share.Link, error) {
	var v []*share.Link
	err := s.db.Select(q.Eq("Path", path), q.Eq("UserID", id)).Find(&v)
	if errors.Is(err, storm.ErrNotFound) {
		return v, fberrors.ErrNotExist
	}

	return v, err
}

func (s shareBackend) Save(l *share.Link) error {
	return s.db.Bolt.Update(func(tx *bolt.Tx) error {
		dbx := s.db.WithTransaction(tx)

		var existing share.Link
		err := dbx.One("Hash", l.Hash, &existing)
		switch {
		case errors.Is(err, storm.ErrNotFound):
		case err != nil:
			return err
		case existing.Expire != 0 && existing.Expire <= time.Now().Unix():
			if err := dbx.DeleteStruct(&share.Link{Hash: existing.Hash}); err != nil && !errors.Is(err, storm.ErrNotFound) {
				return err
			}
		default:
			return fberrors.ErrExist
		}

		return dbx.Save(l)
	})
}

func (s shareBackend) Delete(hash string) error {
	err := s.db.DeleteStruct(&share.Link{Hash: hash})
	if errors.Is(err, storm.ErrNotFound) {
		return nil
	}
	return err
}

func (s shareBackend) DeleteWithPathPrefix(pathPrefix string) error {
	var links []share.Link
	if err := s.db.Prefix("Path", pathPrefix, &links); err != nil {
		return err
	}

	var err error
	for _, link := range links {
		err = errors.Join(err, s.db.DeleteStruct(&share.Link{Hash: link.Hash}))
	}
	return err
}

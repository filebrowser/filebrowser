package bolt

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	fm "github.com/hacdias/filemanager"
)

// ShareStore is a shareable links store.
type ShareStore struct {
	DB *storm.DB
}

// Get gets a Share Link from an hash.
func (s ShareStore) Get(hash string) (*fm.ShareLink, error) {
	var v fm.ShareLink
	err := s.DB.One("Hash", hash, &v)
	if err == storm.ErrNotFound {
		return nil, fm.ErrNotExist
	}

	return &v, err
}

// GetPermanent gets the permanent link from a path.
func (s ShareStore) GetPermanent(path string) (*fm.ShareLink, error) {
	var v fm.ShareLink
	err := s.DB.Select(q.Eq("Path", path), q.Eq("Expires", false)).First(&v)
	if err == storm.ErrNotFound {
		return nil, fm.ErrNotExist
	}

	return &v, err
}

// GetByPath gets all the links for a specific path.
func (s ShareStore) GetByPath(hash string) ([]*fm.ShareLink, error) {
	var v []*fm.ShareLink
	err := s.DB.Find("Path", hash, &v)
	if err == storm.ErrNotFound {
		return v, fm.ErrNotExist
	}

	return v, err
}

// Gets retrieves all the shareable links.
func (s ShareStore) Gets() ([]*fm.ShareLink, error) {
	var v []*fm.ShareLink
	err := s.DB.All(&v)
	if err == storm.ErrNotFound {
		return v, fm.ErrNotExist
	}

	return v, err
}

// Save stores a Share Link on the database.
func (s ShareStore) Save(l *fm.ShareLink) error {
	return s.DB.Save(l)
}

// Delete deletes a Share Link from the database.
func (s ShareStore) Delete(hash string) error {
	return s.DB.DeleteStruct(&fm.ShareLink{Hash: hash})
}

package bolt

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/filebrowser/filebrowser/lib"
)

func (s Backend) GetLinkByHash(hash string) (*lib.ShareLink, error) {
	var v lib.ShareLink
	err := s.DB.One("Hash", hash, &v)
	if err == storm.ErrNotFound {
		return nil, lib.ErrNotExist
	}

	return &v, err
}

func (s Backend) GetLinkPermanent(path string) (*lib.ShareLink, error) {
	var v lib.ShareLink
	err := s.DB.Select(q.Eq("Path", path), q.Eq("Expires", false)).First(&v)
	if err == storm.ErrNotFound {
		return nil, lib.ErrNotExist
	}

	return &v, err
}

func (s Backend) GetLinksByPath(hash string) ([]*lib.ShareLink, error) {
	var v []*lib.ShareLink
	err := s.DB.Find("Path", hash, &v)
	if err == storm.ErrNotFound {
		return v, lib.ErrNotExist
	}

	return v, err
}

func (s Backend) SaveLink(l *lib.ShareLink) error {
	return s.DB.Save(l)
}

func (s Backend) DeleteLink(hash string) error {
	return s.DB.DeleteStruct(&lib.ShareLink{Hash: hash})
}

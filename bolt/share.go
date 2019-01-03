package bolt

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/filebrowser/filebrowser/types"
)

func (s Backend) GetLinkByHash(hash string) (*types.ShareLink, error) {
	var v types.ShareLink
	err := s.DB.One("Hash", hash, &v)
	if err == storm.ErrNotFound {
		return nil, types.ErrNotExist
	}

	return &v, err
}

func (s Backend) GetLinkPermanent(path string) (*types.ShareLink, error) {
	var v types.ShareLink
	err := s.DB.Select(q.Eq("Path", path), q.Eq("Expires", false)).First(&v)
	if err == storm.ErrNotFound {
		return nil, types.ErrNotExist
	}

	return &v, err
}

func (s Backend) GetLinksByPath(hash string) ([]*types.ShareLink, error) {
	var v []*types.ShareLink
	err := s.DB.Find("Path", hash, &v)
	if err == storm.ErrNotFound {
		return v, types.ErrNotExist
	}

	return v, err
}

func (s Backend) SaveLink(l *types.ShareLink) error {
	return s.DB.Save(l)
}

func (s Backend) DeleteLink(hash string) error {
	return s.DB.DeleteStruct(&types.ShareLink{Hash: hash})
}

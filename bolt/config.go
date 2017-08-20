package bolt

import (
	"github.com/asdine/storm"
	fm "github.com/hacdias/filemanager"
)

type ConfigStore struct {
	DB *storm.DB
}

func (c ConfigStore) Get(name string, to interface{}) error {
	err := c.DB.Get("config", name, to)
	if err == storm.ErrNotFound {
		return fm.ErrNotExist
	}

	return err
}

func (c ConfigStore) Save(name string, from interface{}) error {
	return c.DB.Set("config", name, from)
}

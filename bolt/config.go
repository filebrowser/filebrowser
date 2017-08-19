package bolt

import (
	"github.com/asdine/storm"
)

type ConfigStore struct {
	DB *storm.DB
}

func (c ConfigStore) Get(name string, to interface{}) error {
	return c.DB.Get("config", name, to)
}

func (c ConfigStore) Save(name string, from interface{}) error {
	return c.DB.Set("config", name, from)
}

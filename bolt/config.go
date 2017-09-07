package bolt

import (
	"github.com/asdine/storm"
	fm "github.com/hacdias/filemanager"
)

// ConfigStore is a configuration store.
type ConfigStore struct {
	DB *storm.DB
}

// Get gets a configuration from the database to an interface.
func (c ConfigStore) Get(name string, to interface{}) error {
	err := c.DB.Get("config", name, to)
	if err == storm.ErrNotFound {
		return fm.ErrNotExist
	}

	return err
}

// Save saves a configuration from an interface to the database.
func (c ConfigStore) Save(name string, from interface{}) error {
	return c.DB.Set("config", name, from)
}

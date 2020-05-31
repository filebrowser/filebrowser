package bolt

import (
	"github.com/asdine/storm"

	"github.com/filebrowser/filebrowser/v2/errors"
)

func get(db *storm.DB, name string, to interface{}) error {
	err := db.Get("config", name, to)
	if err == storm.ErrNotFound {
		return errors.ErrNotExist
	}

	return err
}

func save(db *storm.DB, name string, from interface{}) error {
	return db.Set("config", name, from)
}

package bolt

import (
	"errors"

	"github.com/asdine/storm/v3"

	fbErrors "github.com/filebrowser/filebrowser/v2/errors"
)

func get(db *storm.DB, name string, to interface{}) error {
	err := db.Get("config", name, to)
	if errors.Is(err, storm.ErrNotFound) {
		return fbErrors.ErrNotExist
	}

	return err
}

func save(db *storm.DB, name string, from interface{}) error {
	return db.Set("config", name, from)
}

package importer

import (
	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
)

// Import imports an old configuration to a newer database.
func Import(oldDB, oldConf, newDB string) error {
	old, err := storm.Open(oldDB)
	if err != nil {
		return err
	}
	defer old.Close()

	new, err := storm.Open(newDB)
	if err != nil {
		return err
	}
	defer new.Close()

	sto, err := bolt.NewStorage(new)
	if err != nil {
		return err
	}

	err = importUsers(old, sto)
	if err != nil {
		return err
	}

	err = importConf(old, oldConf, sto)
	if err != nil {
		return err
	}

	return err
}

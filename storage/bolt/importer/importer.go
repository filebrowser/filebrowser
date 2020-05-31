package importer

import (
	"github.com/asdine/storm"

	"github.com/filebrowser/filebrowser/v2/storage/bolt"
)

// Import imports an old configuration to a newer database.
func Import(oldDBPath, oldConf, newDBPath string) error {
	oldDB, err := storm.Open(oldDBPath)
	if err != nil {
		return err
	}
	defer oldDB.Close()

	newDB, err := storm.Open(newDBPath)
	if err != nil {
		return err
	}
	defer newDB.Close()

	sto, err := bolt.NewStorage(newDB)
	if err != nil {
		return err
	}

	err = importUsers(oldDB, sto)
	if err != nil {
		return err
	}

	err = importConf(oldDB, oldConf, sto)
	if err != nil {
		return err
	}

	return err
}

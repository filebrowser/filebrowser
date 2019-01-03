package bolt

import "github.com/asdine/storm"

// Backend implements types.StorageBackend.
type Backend struct {
	DB *storm.DB
}

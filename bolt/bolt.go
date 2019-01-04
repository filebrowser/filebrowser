package bolt

import "github.com/asdine/storm"

// Backend implements lib.StorageBackend.
type Backend struct {
	DB *storm.DB
}

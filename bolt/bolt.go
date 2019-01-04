package bolt

import "github.com/asdine/storm"

// Backend implements lib.StorageBackend
// using Bolt DB.
type Backend struct {
	DB *storm.DB
}

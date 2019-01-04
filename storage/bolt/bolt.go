package bolt

import "github.com/asdine/storm"

// Backend implements storage.Backend
// using Bolt DB.
type Backend struct {
	DB *storm.DB
}

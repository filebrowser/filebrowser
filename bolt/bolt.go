package bolt

import "github.com/asdine/storm"

func Open(path string) (*storm.DB, error) {
	db, err := storm.Open(path)
	if err != nil {
		return nil, err
	}

	return db, nil
}

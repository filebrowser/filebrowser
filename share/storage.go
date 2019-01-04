package share

// StorageBackend is the interface to implement for a share storage.
type StorageBackend interface {
	GetByHash(hash string) (*Link, error)
	GetPermanent(path string) (*Link, error)
	Gets(path string) ([]*Link, error)
	Save(s *Link) error
	Delete(hash string) error
}

// Storage is a storage.
type Storage struct {
	back StorageBackend
}

// NewStorage creates a share links storage from a backend.
func NewStorage(back StorageBackend) *Storage {
	return &Storage{back: back}
}

// GetByHash wraps a StorageBackend.GetByHash.
func (s *Storage) GetByHash(hash string) (*Link, error) {
	return s.back.GetByHash(hash)
}

// GetPermanent wraps a StorageBackend.GetPermanent
func (s *Storage) GetPermanent(path string) (*Link, error) {
	return s.back.GetPermanent(path)
}

// Gets wraps a StorageBackend.Gets
func (s *Storage) Gets(path string) ([]*Link, error) {
	return s.back.Gets(path)
}

// Save wraps a StorageBackend.Save
func (s *Storage) Save(l *Link) error {
	return s.back.Save(l)
}

// Delete wraps a StorageBackend.Delete
func (s *Storage) Delete(hash string) error {
	return s.back.Delete(hash)
}

package types

// Store is used to persist data.
type Store struct {
	Users  *UsersVerify
	Config *ConfigVerify
	Share  ShareStore
}

// ShareStore is the interface to manage share links.
type ShareStore interface {
	Get(hash string) (*ShareLink, error)
	GetPermanent(path string) (*ShareLink, error)
	GetByPath(path string) ([]*ShareLink, error)
	Gets() ([]*ShareLink, error)
	Save(s *ShareLink) error
	Delete(hash string) error
}

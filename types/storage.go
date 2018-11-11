package types

// Store is used to persist data.
type Store struct {
	Users  UsersStore
	Config ConfigStore
	Share  ShareStore
}

// UsersStore is used to manage users relativey to a data storage.
type UsersStore interface {
	Get(id uint) (*User, error)
	GetByUsername(username string) (*User, error)
	Gets() ([]*User, error)
	Save(u *User) error
	Update(u *User, fields ...string) error
	Delete(id uint) error
}

// ConfigStore is used to manage configurations relativey to a data storage.
type ConfigStore interface {
	Get(name string, to interface{}) error
	Save(name string, from interface{}) error
	GetSettings() (*Settings, error)
	SaveSettings(*Settings) error
	SaveRunner(*Runner) error
	GetRunner() (*Runner, error)
	GetAuther(AuthMethod) (Auther, error)
	SaveAuther(Auther) error
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

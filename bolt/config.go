package bolt

import (
	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/auth"
	"github.com/filebrowser/filebrowser/types"
)

// ConfigStore is a configuration store.
type ConfigStore struct {
	DB *storm.DB
	Users types.UsersStore
}

// Get gets a configuration from the database to an interface.
func (c ConfigStore) Get(name string, to interface{}) error {
	err := c.DB.Get("config", name, to)
	if err == storm.ErrNotFound {
		return types.ErrNotExist
	}

	return err
}

// Save saves a configuration from an interface to the database.
func (c ConfigStore) Save(name string, from interface{}) error {
	return c.DB.Set("config", name, from)
}

// GetSettings is an helper method to get a settings object.
func (c ConfigStore) GetSettings() (*types.Settings, error) {
	settings := &types.Settings{}
	return settings, c.Get("settings", settings)
}

// SaveSettings is an helper method to set the settings object
func (c ConfigStore) SaveSettings(s *types.Settings) error {
	return c.Save("settings", s)
}

// GetRunner is an helper method to get a runner object.
func (c ConfigStore) GetRunner() (*types.Runner, error) {
	runner := &types.Runner{}
	return runner, c.Get("runner", runner)
}

// SaveRunner is an helper method to set the runner object
func (c ConfigStore) SaveRunner (r *types.Runner) error {
	return c.Save("runner", r)
}

// GetAuther is an helper method to get an auther object.
func (c ConfigStore) GetAuther(t types.AuthMethod) (types.Auther, error) {
	if t == auth.MethodJSONAuth {
		auther := auth.JSONAuth{}
		if err := c.Get("auther", &auther); err != nil {
			return nil, err
		}
		auther.Store = &UsersStore{DB: c.DB}
		return &auther, nil
	}

	if t == auth.MethodProxyAuth {
		auther := auth.ProxyAuth{}
		if err := c.Get("auther", &auther); err != nil {
			return nil, err
		}
		return &auther, nil
	}

	if t == auth.MethodNoAuth {
		auther := auth.NoAuth{Store: c.Users}
		if err := c.Get("auther", &auther); err != nil {
			return nil, err
		}
		return &auther, nil
	}

	return nil, types.ErrInvalidAuthMethod
}

// SaveAuther is an helper method to set the auther object
func (c ConfigStore) SaveAuther(a types.Auther) error {
	return c.Save("auther", a)
}

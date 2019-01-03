package bolt

import (
	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/auth"
	"github.com/filebrowser/filebrowser/types"
)

func (b Backend) get(name string, to interface{}) error {
	err := b.DB.Get("config", name, to)
	if err == storm.ErrNotFound {
		return types.ErrNotExist
	}

	return err
}

func (b Backend) save(name string, from interface{}) error {
	return b.DB.Set("config", name, from)
}

func (b Backend) GetSettings() (*types.Settings, error) {
	settings := &types.Settings{}
	return settings, b.get("settings", settings)
}

func (b Backend) SaveSettings(s *types.Settings) error {
	return b.save("settings", s)
}

func (b Backend) GetAuther(t types.AuthMethod) (types.Auther, error) {
	var auther types.Auther

	switch t {
	case auth.MethodJSONAuth:
		auther = &auth.JSONAuth{}
	case auth.MethodProxyAuth:
		auther = &auth.ProxyAuth{}
	case auth.MethodNoAuth:
		auther = &auth.NoAuth{}
	default:
		return nil, types.ErrInvalidAuthMethod
	}

	return auther, b.get("auther", auther)
}

func (b Backend) SaveAuther(a types.Auther) error {
	return b.save("auther", a)
}

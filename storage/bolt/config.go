package bolt

import (
	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/auth"
	
)

func (b Backend) get(name string, to interface{}) error {
	err := b.DB.Get("config", name, to)
	if err == storm.ErrNotFound {
		return lib.ErrNotExist
	}

	return err
}

func (b Backend) save(name string, from interface{}) error {
	return b.DB.Set("config", name, from)
}

func (b Backend) GetSettings() (*settings.Settings, error) {
	settings := &settings.Settings{}
	return settings, b.get("settings", settings)
}

func (b Backend) SaveSettings(s *settings.Settings) error {
	return b.save("settings", s)
}

func (b Backend) GetAuther(t lib.AuthMethod) (lib.Auther, error) {
	var auther lib.Auther

	switch t {
	case auth.MethodJSONAuth:
		auther = &auth.JSONAuth{}
	case auth.MethodProxyAuth:
		auther = &auth.ProxyAuth{}
	case auth.MethodNoAuth:
		auther = &auth.NoAuth{}
	default:
		return nil, lib.ErrInvalidAuthMethod
	}

	return auther, b.get("auther", auther)
}

func (b Backend) SaveAuther(a lib.Auther) error {
	return b.save("auther", a)
}

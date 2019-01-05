package bolt

import (
	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/v2/settings"
)

type settingsBackend struct {
	db *storm.DB
}

func (s settingsBackend) Get() (*settings.Settings, error) {
	settings := &settings.Settings{}
	return settings, get(s.db, "settings", settings)
}

func (s settingsBackend) Save(settings *settings.Settings) error {
	return save(s.db, "settings", settings)
}

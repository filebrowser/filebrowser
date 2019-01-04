package settings

import (
	"strings"

	"github.com/filebrowser/filebrowser/errors"
	"github.com/filebrowser/filebrowser/rules"
	"github.com/filebrowser/filebrowser/users"
)

// StorageBackend is a settings storage backend.
type StorageBackend interface {
	Get() (*Settings, error)
	Save(*Settings) error
}

// Storage is a settings storage.
type Storage struct {
	back StorageBackend
}

// NewStorage creates a settings storage from a backend.
func NewStorage(back StorageBackend) *Storage {
	return &Storage{back: back}
}

// Get returns the settings for the current instance.
func (s *Storage) Get() (*Settings, error) {
	return s.back.Get()
}

var defaultEvents = []string{
	"save",
	"copy",
	"rename",
	"upload",
	"delete",
}

// Save saves the settings for the current instance.
func (s *Storage) Save(set *Settings) error {
	set.BaseURL = strings.TrimSuffix(set.BaseURL, "/")

	if len(set.Key) == 0 {
		return errors.ErrEmptyKey
	}

	if set.Defaults.Locale == "" {
		set.Defaults.Locale = "en"
	}

	if set.Defaults.Commands == nil {
		set.Defaults.Commands = []string{}
	}

	if set.Defaults.ViewMode == "" {
		set.Defaults.ViewMode = users.MosaicViewMode
	}

	if set.Rules == nil {
		set.Rules = []rules.Rule{}
	}

	if set.Shell == nil {
		set.Shell = []string{}
	}

	if set.Commands == nil {
		set.Commands = map[string][]string{}
	}

	for _, event := range defaultEvents {
		if _, ok := set.Commands["before_"+event]; !ok {
			set.Commands["before_"+event] = []string{}
		}

		if _, ok := set.Commands["after_"+event]; !ok {
			set.Commands["after_"+event] = []string{}
		}
	}

	err := s.back.Save(set)
	if err != nil {
		return err
	}

	return nil
}

package types

import "strings"

// ConfigStore is used to manage configurations relativey to a data storage.
type ConfigStore interface {
	GetSettings() (*Settings, error)
	SaveSettings(*Settings) error
	GetAuther(AuthMethod) (Auther, error)
	SaveAuther(Auther) error
}

// ConfigVerify wraps a ConfigStore and makes the verifications needed.
type ConfigVerify struct {
	Store ConfigStore
}

// GetSettings wraps a ConfigStore.GetSettings
func (v ConfigVerify) GetSettings() (*Settings, error) {
	return v.Store.GetSettings()
}

// SaveSettings wraps a ConfigStore.SaveSettings
func (v ConfigVerify) SaveSettings(s *Settings) error {
	s.BaseURL = strings.TrimSuffix(s.BaseURL, "/")

	if len(s.Key) == 0 {
		return ErrEmptyKey
	}

	if s.Defaults.Locale == "" {
		s.Defaults.Locale = "en"
	}

	if s.Defaults.Commands == nil {
		s.Defaults.Commands = []string{}
	}

	if s.Defaults.ViewMode == "" {
		s.Defaults.ViewMode = MosaicViewMode
	}

	if s.Rules == nil {
		s.Rules = []Rule{}
	}

	if s.Shell == nil {
		s.Shell = []string{}
	}

	if s.Commands == nil {
		s.Commands = map[string][]string{}
	}

	for _, event := range defaultEvents {
		if _, ok := s.Commands["before_"+event]; !ok {
			s.Commands["before_"+event] = []string{}
		}

		if _, ok := s.Commands["after_"+event]; !ok {
			s.Commands["after_"+event] = []string{}
		}
	}

	return v.Store.SaveSettings(s)
}

// GetAuther wraps a ConfigStore.GetAuther
func (v ConfigVerify) GetAuther(t AuthMethod) (Auther, error) {
	return v.Store.GetAuther(t)
}

// SaveAuther wraps a ConfigStore.SaveAuther
func (v ConfigVerify) SaveAuther(a Auther) error {
	return v.Store.SaveAuther(a)
}

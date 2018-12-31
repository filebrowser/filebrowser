package types

import "strings"

// ConfigStore is used to manage configurations relativey to a data storage.
type ConfigStore interface {
	GetSettings() (*Settings, error)
	SaveSettings(*Settings) error
	SaveRunner(*Runner) error
	GetRunner() (*Runner, error)
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

	return v.Store.SaveSettings(s)
}

// GetRunner wraps a ConfigStore.GetRunner
func (v ConfigVerify) GetRunner() (*Runner, error) {
	return v.Store.GetRunner()
}

// SaveRunner wraps a ConfigStore.SaveRunner
func (v ConfigVerify) SaveRunner(r *Runner) error {
	if r.Commands == nil {
		r.Commands = map[string][]string{}
	}

	for _, event := range defaultEvents {
		if _, ok := r.Commands["before_"+event]; !ok {
			r.Commands["before_"+event] = []string{}
		}

		if _, ok := r.Commands["after_"+event]; !ok {
			r.Commands["after_"+event] = []string{}
		}
	}

	return v.Store.SaveRunner(r)
}

// GetAuther wraps a ConfigStore.GetAuther
func (v ConfigVerify) GetAuther(t AuthMethod) (Auther, error) {
	return v.Store.GetAuther(t)
}

// SaveAuther wraps a ConfigStore.SaveAuther
func (v ConfigVerify) SaveAuther(a Auther) error {
	return v.Store.SaveAuther(a)
}

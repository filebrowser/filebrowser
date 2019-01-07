package settings

import "github.com/filebrowser/filebrowser/v2/rules"

// AuthMethod describes an authentication method.
type AuthMethod string

// Settings contain the main settings of the application.
type Settings struct {
	Runtime    map[string]string   `json:"runtime"`
	Signup     bool                `json:"signup"`
	Defaults   UserDefaults        `json:"defaults"`
	AuthMethod AuthMethod          `json:"authMethod"`
	Branding   Branding            `json:"branding"`
	Commands   map[string][]string `json:"commands"`
	Shell      []string            `json:"shell"`
	Rules      []rules.Rule        `json:"rules"`
}

// GetRules implements rules.Provider.
func (s *Settings) GetRules() []rules.Rule {
	return s.Rules
}

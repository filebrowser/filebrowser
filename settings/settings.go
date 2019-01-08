package settings

import "github.com/filebrowser/filebrowser/v2/rules"

// AuthMethod describes an authentication method.
type AuthMethod string

// Settings contain the main settings of the application.
type Settings struct {
	Key        []byte              `json:"key"`
	Signup     bool                `json:"signup"`
	Defaults   UserDefaults        `json:"defaults"`
	AuthMethod AuthMethod          `json:"authMethod"`
	Branding   Branding            `json:"branding"`
	Commands   map[string][]string `json:"commands"`
	Shell      []string            `json:"shell"`
	Rules      []rules.Rule        `json:"rules"`
}

// Server specific settings.
type Server struct {
	Root    string `json:"root"`
	BaseURL string `json:"baseURL"`
	TLSKey  string `json:"tlsKey"`
	TLSCert string `json:"tlsCert"`
	Port    string `json:"port"`
	Address string `json:"address"`
	Log     string `json:"log"`
}

// GetRules implements rules.Provider.
func (s *Settings) GetRules() []rules.Rule {
	return s.Rules
}

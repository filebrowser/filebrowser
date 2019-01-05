package settings

import "github.com/filebrowser/filebrowser/rules"

// AuthMethod describes an authentication method.
type AuthMethod string

// Settings contain the main settings of the application.
type Settings struct {
	Key        []byte              `json:"key"`
	BaseURL    string              `json:"baseURL"`
	Log        string              `json:"log"`
	Server     Server              `json:"server"`
	Signup     bool                `json:"signup"`
	Defaults   UserDefaults        `json:"defaults"`
	AuthMethod AuthMethod          `json:"authMethod"`
	Branding   Branding            `json:"branding"`
	Commands   map[string][]string `json:"commands"`
	Shell      []string            `json:"shell"`
	Rules      []rules.Rule        `json:"rules"` // TODO: use this add to cli
}

// Server settings.
type Server struct {
	Port    int    `json:"port"`
	Address string `json:"address"`
	TLSCert string `json:"tlsCert"`
	TLSKey  string `json:"tlsKey"`
}

// GetRules implements rules.Provider.
func (s *Settings) GetRules() []rules.Rule {
	return s.Rules
}

package settings

import "github.com/filebrowser/filebrowser/v2/rules"

// AuthMethod describes an authentication method.
type AuthMethod string

// Settings contain the main settings of the application.
type Settings struct {
	Key        []byte              `json:"key"`
	BaseURL    string              `json:"baseURL"`
	Log        string              `json:"log"`
	Scope      string              `json:"scope"`
	Server     Server              `json:"server"`
	Signup     bool                `json:"signup"`
	Defaults   UserDefaults        `json:"defaults"`
	AuthMethod AuthMethod          `json:"authMethod"`
	Branding   Branding            `json:"branding"`
	Commands   map[string][]string `json:"commands"`
	Shell      []string            `json:"shell"`
	Rules      []rules.Rule        `json:"rules"`
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

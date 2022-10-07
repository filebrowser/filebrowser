package settings

import (
	"crypto/rand"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/rules"
)

const DefaultUsersHomeBasePath = "/users"

// AuthMethod describes an authentication method.
type AuthMethod string

// Settings contain the main settings of the application.
type Settings struct {
	Key                 []byte              `json:"key"`
	Signup              bool                `json:"signup"`
	CreateUserDir       bool                `json:"createUserDir"`
	UserHomeBasePath    string              `json:"userHomeBasePath"`
	Defaults            UserDefaults        `json:"defaults"`
	AuthMethod          AuthMethod          `json:"authMethod"`
	Branding            Branding            `json:"branding"`
	Commands            map[string][]string `json:"commands"`
	Shell               []string            `json:"shell"`
	Rules               []rules.Rule        `json:"rules"`
	TokenExpirationTime Duration            `json:"tokenExpirationTime"` // 0 is treated as 2 Hours
}

// GetRules implements rules.Provider.
func (s *Settings) GetRules() []rules.Rule {
	return s.Rules
}

// Server specific settings.
type Server struct {
	Root                  string `json:"root"`
	BaseURL               string `json:"baseURL"`
	Socket                string `json:"socket"`
	TLSKey                string `json:"tlsKey"`
	TLSCert               string `json:"tlsCert"`
	Port                  string `json:"port"`
	Address               string `json:"address"`
	Log                   string `json:"log"`
	EnableThumbnails      bool   `json:"enableThumbnails"`
	ResizePreview         bool   `json:"resizePreview"`
	EnableExec            bool   `json:"enableExec"`
	TypeDetectionByHeader bool   `json:"typeDetectionByHeader"`
	AuthHook              string `json:"authHook"`
}

// Clean cleans any variables that might need cleaning.
func (s *Server) Clean() {
	s.BaseURL = strings.TrimSuffix(s.BaseURL, "/")
}

// GenerateKey generates a key of 512 bits.
func GenerateKey() ([]byte, error) {
	b := make([]byte, 64) //nolint:gomnd
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

type Duration time.Duration // support json Marshal/Unmarshal for time.Duration

func (dur Duration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(time.Duration(dur).String())), nil
}

func (dur *Duration) UnmarshalJSON(data []byte) error {
	var dStr string
	err := json.Unmarshal(data, &dStr)
	if err != nil {
		return err
	}
	if dStr == "" {
		*dur = 0 // zero value
		return nil
	}
	d, err := time.ParseDuration(dStr)
	if err != nil {
		return err
	}
	*dur = Duration(d)
	return nil
}

func (dur Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(dur).String(), nil
}

func (dur *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var dStr string
	err := unmarshal(&dStr)
	if err != nil {
		return err
	}
	if dStr == "" {
		*dur = 0 // zero value
		return nil
	}
	d, err := time.ParseDuration(dStr)
	if err != nil {
		return err
	}
	*dur = Duration(d)
	return nil
}

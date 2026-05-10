package settings

import (
	"crypto/rand"
	"io/fs"
	"log"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/rules"
)

const DefaultUsersHomeBasePath = "/users"
const DefaultLogoutPage = "/login"
const DefaultMinimumPasswordLength = 12
const DefaultFileMode = 0640
const DefaultDirMode = 0750

// AuthMethod describes an authentication method.
type AuthMethod string

// DefaultHaasPort is the Waveshare RS-232↔TCP bridge port the Pi opens
// to drip-feed and query the Haas. Override per-instance via the Machine
// settings tab.
const DefaultHaasPort = 4196

// Cnc holds the per-instance machine integration config that the
// /api/cnc/* endpoints read and the Machine settings tab edits.
//
// Stored on the same Settings record (single Bolt JSON blob) — adding a
// nested struct is forward-compatible: pre-existing DBs decode it as the
// zero value, and Storage.Get fills sensible defaults.
type Cnc struct {
	HaasHost  string `json:"haasHost"`
	HaasPort  int    `json:"haasPort"`
	CameraURL string `json:"cameraUrl"`
	// MachineToken is an opaque random secret used as a bearer token
	// for server-to-server access to /api/cnc/state and /api/cnc/qcode
	// (Home Assistant scripts, monitoring, etc). Originally minted for
	// the haas-dashboard project (now subsumed by filebrowser-NC); the
	// mechanism stays because it's useful for any external consumer.
	// Empty until the admin clicks "Regenerate" in the UI.
	MachineToken string `json:"machineToken"`
}

// Settings contain the main settings of the application.
type Settings struct {
	Key                   []byte              `json:"key"`
	Signup                bool                `json:"signup"`
	HideLoginButton       bool                `json:"hideLoginButton"`
	CreateUserDir         bool                `json:"createUserDir"`
	UserHomeBasePath      string              `json:"userHomeBasePath"`
	Defaults              UserDefaults        `json:"defaults"`
	AuthMethod            AuthMethod          `json:"authMethod"`
	LogoutPage            string              `json:"logoutPage"`
	Branding              Branding            `json:"branding"`
	Tus                   Tus                 `json:"tus"`
	Commands              map[string][]string `json:"commands"`
	Shell                 []string            `json:"shell"`
	Rules                 []rules.Rule        `json:"rules"`
	MinimumPasswordLength uint                `json:"minimumPasswordLength"`
	FileMode              fs.FileMode         `json:"fileMode"`
	DirMode               fs.FileMode         `json:"dirMode"`
	HideDotfiles          bool                `json:"hideDotfiles"`
	Cnc                   Cnc                 `json:"cnc"`
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
	ImageResolutionCal    bool   `json:"imageResolutionCalculation"`
	AuthHook              string `json:"authHook"`
	TokenExpirationTime   string `json:"tokenExpirationTime"`
}

// Clean cleans any variables that might need cleaning.
func (s *Server) Clean() {
	s.BaseURL = strings.TrimSuffix(s.BaseURL, "/")
}

func (s *Server) GetTokenExpirationTime(fallback time.Duration) time.Duration {
	if s.TokenExpirationTime == "" {
		return fallback
	}

	duration, err := time.ParseDuration(s.TokenExpirationTime)
	if err != nil {
		log.Printf("[WARN] Failed to parse tokenExpirationTime: %v", err)
		return fallback
	}
	return duration
}

// GenerateKey generates a key of 512 bits.
func GenerateKey() ([]byte, error) {
	b := make([]byte, 64)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

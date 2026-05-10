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
// Multi-machine support (2026-05-10): Machines is the canonical list.
// Legacy fields (HaasHost/HaasPort/CameraURL) are kept on the struct
// for one-time migration from pre-multi-machine DBs and are NOT read
// by new code. EnsureMigrated() folds them into Machines[0] on first
// boot of the new binary.
type Cnc struct {
	// Machines is the canonical machine list. First entry is the
	// default for any /api/cnc/* call without an explicit machine_id.
	Machines []Machine `json:"machines"`

	// MachineToken is the long-lived bearer used by external services
	// (HA, monitoring, custom dashboards) to call /api/cnc/state or
	// /api/cnc/qcode without a filebrowser session. Global, not
	// per-machine — one token covers all machines under this install.
	MachineToken string `json:"machineToken"`

	// ── Legacy fields (deprecated; migrated into Machines[0]) ──
	HaasHost  string `json:"haasHost,omitempty"`
	HaasPort  int    `json:"haasPort,omitempty"`
	CameraURL string `json:"cameraUrl,omitempty"`
}

// Machine is one configured CNC controller. All Haas-shaped today;
// brand abstraction can land later without breaking the schema (a
// Brand field default-empties to "haas").
type Machine struct {
	// ID is stable across renames. Generated on creation; never
	// edited. The default-machine selector uses this.
	ID string `json:"id"`
	// Name is the operator-facing label, freely editable.
	Name string `json:"name"`
	// Host:Port is the Waveshare RS-232↔TCP bridge.
	Host string `json:"host"`
	Port int    `json:"port"`
	// CameraURL is optional; HLS/snapshot/RTSP-hint per the existing
	// camera tile dispatch.
	CameraURL string `json:"cameraUrl,omitempty"`
}

// EnsureMigrated folds legacy single-machine fields into Machines[0]
// if Machines is empty. Idempotent; safe to call on every Settings
// load. Returns true when a migration actually happened so the caller
// can persist.
func (c *Cnc) EnsureMigrated() bool {
	if len(c.Machines) > 0 {
		return false
	}
	if c.HaasHost == "" && c.HaasPort == 0 && c.CameraURL == "" {
		// Brand-new install — no machines yet, nothing to migrate.
		return false
	}
	port := c.HaasPort
	if port == 0 {
		port = DefaultHaasPort
	}
	c.Machines = []Machine{{
		ID:        "primary",
		Name:      "Machine 1",
		Host:      c.HaasHost,
		Port:      port,
		CameraURL: c.CameraURL,
	}}
	return true
}

// MachineByID returns the matching Machine and true, or zero + false.
func (c *Cnc) MachineByID(id string) (Machine, bool) {
	for _, m := range c.Machines {
		if m.ID == id {
			return m, true
		}
	}
	return Machine{}, false
}

// DefaultMachineID returns the ID treated as the default when an
// API call doesn't specify one. First entry of Machines, or "" if
// no machines are configured.
func (c *Cnc) DefaultMachineID() string {
	if len(c.Machines) == 0 {
		return ""
	}
	return c.Machines[0].ID
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

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

// Machine is one configured CNC controller.
type Machine struct {
	// ID is stable across renames. Generated on creation; never
	// edited. The default-machine selector uses this.
	ID string `json:"id"`
	// Name is the operator-facing label, freely editable.
	Name string `json:"name"`
	// Brand identifies the controller family. Only "haas" is wired
	// today; the field exists so a per-brand send protocol / state
	// dialect can be slotted in without a settings-schema migration.
	// Empty values normalize to "haas" on save.
	Brand string `json:"brand,omitempty"`
	// Host:Port is the Waveshare RS-232↔TCP bridge.
	Host string `json:"host"`
	Port int    `json:"port"`
	// ToolSlots is the magazine capacity for tool-table reads. Operators
	// set this to their machine's actual slot count (e.g. 20 for a
	// 20-pocket carousel) so reads cover the whole magazine without
	// probing the unreachable upper range. 0 falls back to
	// DefaultToolSlots.
	ToolSlots int `json:"toolSlots,omitempty"`
	// CameraURL is optional. CameraType picks the rendering path.
	CameraURL string `json:"cameraUrl,omitempty"`
	// CameraType is one of "auto" / "hls" / "mjpeg" / "iframe" /
	// "none". Empty normalizes to "auto" (legacy URL-suffix dispatch).
	// "iframe" is required for UniFi Protect / Reolink web UI URLs
	// since browsers cannot play raw RTSP/RTSPS.
	CameraType string `json:"cameraType,omitempty"`
	// RequirePreflight, when true, refuses /api/cnc/start if the
	// preflight comparison flags any tools as missing / empty pocket
	// for the program's T-codes. The wizard already soft-warns; this
	// flips the check to a hard server-side gate so an operator can't
	// "I know what I'm doing" past a missing tool. Off by default —
	// the operator-side controller prep is still on them.
	RequirePreflight bool `json:"requirePreflight,omitempty"`
	// AxesEnabled controls which axes the /machine dashboard renders
	// rows for. X, Y, Z are always present; A, B, C are optional —
	// some machines have a 4th or 5th axis, most don't. Stored as a
	// list of uppercase letters; empty / unset defaults to X+Y+Z.
	AxesEnabled []string `json:"axesEnabled,omitempty"`
	// PositionToleranceIn is the in-inches drift between commanded
	// and machine position that flips the dashboard's Δ-CMD readout
	// from green to amber. 0 / unset falls back to 0.001".
	PositionToleranceIn float64 `json:"positionToleranceIn,omitempty"`
	// DPRNTCapture enables a per-write scavenger on the streaming
	// socket that surfaces Haas DPRNT macro output as live events
	// on the WS feed. Off by default — adds a 1-2ms per-line read
	// during a job. Operators using DPRNT for in-cycle probing /
	// measurement output should turn it on.
	DPRNTCapture bool `json:"dprntCapture,omitempty"`
}

// EffectiveAxes returns the axes to render for this machine. Defaults
// to X/Y/Z when AxesEnabled is empty. Letters are uppercased and
// deduped; A/B/C are accepted but anything else is dropped.
func (m Machine) EffectiveAxes() []string {
	if len(m.AxesEnabled) == 0 {
		return []string{"X", "Y", "Z"}
	}
	allow := map[string]bool{"X": true, "Y": true, "Z": true, "A": true, "B": true, "C": true}
	seen := map[string]bool{}
	out := make([]string, 0, len(m.AxesEnabled))
	for _, a := range m.AxesEnabled {
		u := strings.ToUpper(strings.TrimSpace(a))
		if !allow[u] || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	if len(out) == 0 {
		return []string{"X", "Y", "Z"}
	}
	return out
}

// EffectivePositionTolerance returns the green/amber threshold in
// inches. Default 0.001" — same as the Haas's own position-drift
// tolerance for most setups.
func (m Machine) EffectivePositionTolerance() float64 {
	if m.PositionToleranceIn <= 0 {
		return 0.001
	}
	return m.PositionToleranceIn
}

// MachineBrandHaas is the only brand wired into the streamer/aggregator
// today. Other brands round-trip through settings but no protocol code
// reads them yet.
const MachineBrandHaas = "haas"

// DefaultToolSlots is the fallback magazine size when a machine's
// ToolSlots is 0. 30 covers most older Haas mills; operators with
// 200-slot tombstones should set ToolSlots explicitly.
const DefaultToolSlots = 30

// EffectiveToolSlots returns the machine's ToolSlots clamped to the
// valid Haas tool-table range. 0 means use the default.
func (m Machine) EffectiveToolSlots() int {
	if m.ToolSlots <= 0 {
		return DefaultToolSlots
	}
	if m.ToolSlots > 200 {
		return 200
	}
	return m.ToolSlots
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
		ID:         "primary",
		Name:       "Machine 1",
		Brand:      MachineBrandHaas,
		Host:       c.HaasHost,
		Port:       port,
		CameraURL:  c.CameraURL,
		CameraType: "auto",
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

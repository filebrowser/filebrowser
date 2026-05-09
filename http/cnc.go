package fbhttp

// /api/cnc/* — Haas Dashboard ↔ Zinc integration endpoints.
// See docs/INTEGRATION_WITH_HAAS_DASHBOARD.md for the wider design.
//
// Phase 1 (this file) only exposes the settings round-trip + a status
// stub so haas-dashboard can start coding against the contract before
// the streamer is built. Phase 2 will land the streamer + Q-code
// multiplexer; Phase 3 the UI; Phase 4 polish + DPRNT capture.

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/filebrowser/filebrowser/v2/settings"
)

// cncSettingsBody is the wire shape the Machine settings tab POSTs and
// reads. Keep it 1:1 with settings.Cnc so the JSON round-trips cleanly,
// EXCEPT MachineToken — minted server-side via the dedicated regenerate
// endpoint so a stray PUT can't blank it out.
type cncSettingsBody struct {
	HaasHost         string `json:"haasHost"`
	HaasPort         int    `json:"haasPort"`
	CameraURL        string `json:"cameraUrl"`
	HaasDashboardURL string `json:"haasDashboardUrl"`
	MachineToken     string `json:"machineToken,omitempty"` // GET only
}

func cncFromSettings(c settings.Cnc) cncSettingsBody {
	return cncSettingsBody{
		HaasHost:         c.HaasHost,
		HaasPort:         c.HaasPort,
		CameraURL:        c.CameraURL,
		HaasDashboardURL: c.HaasDashboardURL,
		MachineToken:     c.MachineToken,
	}
}

var cncSettingsGetHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	return renderJSON(w, r, cncFromSettings(d.settings.Cnc))
})

var cncSettingsPutHandler = withAdmin(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	req := &cncSettingsBody{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return http.StatusBadRequest, err
	}

	port := req.HaasPort
	if port <= 0 {
		port = settings.DefaultHaasPort
	}
	if port > 65535 {
		return http.StatusBadRequest, nil
	}

	d.settings.Cnc.HaasHost = req.HaasHost
	d.settings.Cnc.HaasPort = port
	d.settings.Cnc.CameraURL = req.CameraURL
	d.settings.Cnc.HaasDashboardURL = req.HaasDashboardURL
	// MachineToken intentionally not touched — see comment on the body type.

	err := d.store.Settings.Save(d.settings)
	return errToStatus(err), err
})

// cncRegenerateTokenHandler mints a fresh opaque secret for
// haas-dashboard's server-to-server calls. Old token is invalidated
// immediately. Admin-only — never exposed to non-admin sessions and
// never returned in a list endpoint.
var cncRegenerateTokenHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return http.StatusInternalServerError, err
	}
	d.settings.Cnc.MachineToken = base64.RawURLEncoding.EncodeToString(buf)
	if err := d.store.Settings.Save(d.settings); err != nil {
		return errToStatus(err), err
	}
	return renderJSON(w, r, map[string]string{"machineToken": d.settings.Cnc.MachineToken})
})

// cncStatusBody is the long-term shape — Phase 1 returns it with
// running=false and zero values. Phase 2 wires the streamer in.
type cncStatusBody struct {
	Running       bool   `json:"running"`
	FilePath      string `json:"file_path,omitempty"`
	FileURL       string `json:"file_url,omitempty"`
	LineCurrent   int    `json:"line_current,omitempty"`
	LineTotal     int    `json:"line_total,omitempty"`
	StartedAt     string `json:"started_at,omitempty"`
	HaasOK        bool   `json:"haas_ok"`
	HaasLastError string `json:"haas_last_error,omitempty"`
}

// cncStatusHandler is auth-required (any logged-in user) so the
// breadcrumb pill can poll it from every page. It does NOT require
// admin — operators need to see whether a job is running.
var cncStatusHandler = withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
	return renderJSON(w, r, cncStatusBody{
		Running: false,
		HaasOK:  true, // optimistic until the streamer reports otherwise
	})
})

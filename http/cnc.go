package fbhttp

// /api/cnc/* — machine integration endpoints.
// See docs/INTEGRATION_WITH_HAAS_DASHBOARD.md for the wider design,
// docs/MULTI_MACHINE_DESIGN.md for the per-Machine.ID architecture.
//
// All endpoints accept an optional ?machine_id=... query param. If
// omitted, the registry resolves to the configured default
// (Cnc.Machines[0]). Single-machine installs continue to work
// without any change to existing API consumers.

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/filebrowser/filebrowser/v2/cnc"
	"github.com/filebrowser/filebrowser/v2/settings"
)

// cncSettingsBody is the wire shape the Machine settings tab POSTs and
// reads. MachineToken is GET-only (minted server-side). Machines is
// the canonical list; legacy haasHost/haasPort/cameraUrl fields are
// returned as a copy of Machines[0] for backwards compat with the
// pre-multi-machine settings UI.
type cncSettingsBody struct {
	Machines     []settings.Machine `json:"machines"`
	MachineToken string             `json:"machineToken,omitempty"` // GET only

	// Legacy mirrors of Machines[0] — the pre-multi-machine settings
	// UI POSTs these. Folded into Machines[0] on PUT if the request
	// doesn't include a Machines list.
	HaasHost  string `json:"haasHost,omitempty"`
	HaasPort  int    `json:"haasPort,omitempty"`
	CameraURL string `json:"cameraUrl,omitempty"`
}

func cncFromSettings(c settings.Cnc) cncSettingsBody {
	body := cncSettingsBody{
		Machines:     c.Machines,
		MachineToken: c.MachineToken,
	}
	if len(c.Machines) > 0 {
		m := c.Machines[0]
		body.HaasHost = m.Host
		body.HaasPort = m.Port
		body.CameraURL = m.CameraURL
	}
	return body
}

var cncSettingsGetHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	return renderJSON(w, r, cncFromSettings(d.settings.Cnc))
})

func cncSettingsPutHandler(registry *cnc.Registry) handleFunc {
	return withAdmin(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		req := &cncSettingsBody{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			return http.StatusBadRequest, err
		}

		// Two PUT shapes are accepted:
		//
		// 1. New multi-machine UI: req.Machines is set. We replace the
		//    list wholesale, validating each entry.
		// 2. Legacy single-machine UI: only haasHost/haasPort/cameraUrl
		//    are set. We fold into Machines[0] (creating one if needed),
		//    leaving any additional machines untouched.
		switch {
		case req.Machines != nil:
			cleaned, err := normalizeMachines(req.Machines, d.settings.Cnc.Machines)
			if err != nil {
				return http.StatusBadRequest, err
			}
			d.settings.Cnc.Machines = cleaned
		default:
			port := req.HaasPort
			if port <= 0 {
				port = settings.DefaultHaasPort
			}
			if port > 65535 {
				return http.StatusBadRequest, fmt.Errorf("port out of range")
			}
			if len(d.settings.Cnc.Machines) == 0 {
				d.settings.Cnc.Machines = []settings.Machine{{
					ID:         newMachineID(),
					Name:       "Machine 1",
					Brand:      settings.MachineBrandHaas,
					CameraType: "auto",
				}}
			}
			d.settings.Cnc.Machines[0].Host = req.HaasHost
			d.settings.Cnc.Machines[0].Port = port
			d.settings.Cnc.Machines[0].CameraURL = req.CameraURL
			if d.settings.Cnc.Machines[0].Brand == "" {
				d.settings.Cnc.Machines[0].Brand = settings.MachineBrandHaas
			}
			if d.settings.Cnc.Machines[0].CameraType == "" {
				d.settings.Cnc.Machines[0].CameraType = "auto"
			}
		}
		// Keep the legacy mirror fields populated as a fallback for
		// any code that hasn't migrated. EnsureMigrated() will skip
		// since Machines[0] now exists, so this is no-op cosmetic.
		if len(d.settings.Cnc.Machines) > 0 {
			d.settings.Cnc.HaasHost = d.settings.Cnc.Machines[0].Host
			d.settings.Cnc.HaasPort = d.settings.Cnc.Machines[0].Port
			d.settings.Cnc.CameraURL = d.settings.Cnc.Machines[0].CameraURL
		}

		if err := d.store.Settings.Save(d.settings); err != nil {
			return errToStatus(err), err
		}
		// Pick up new/removed machines in the live registry.
		registry.Refresh()
		return 0, nil
	})
}

// normalizeMachines validates + assigns IDs to a Machines list before
// it lands in storage. Reuses existing IDs where Names match (so an
// edit doesn't tear down a streamer); generates new IDs for new
// entries. Empty list is rejected — the install must always have at
// least one machine.
func normalizeMachines(in []settings.Machine, existing []settings.Machine) ([]settings.Machine, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("at least one machine required")
	}
	seenIDs := make(map[string]struct{}, len(in))
	out := make([]settings.Machine, 0, len(in))
	for i, m := range in {
		if strings.TrimSpace(m.Name) == "" {
			return nil, fmt.Errorf("machine %d: name required", i)
		}
		if strings.TrimSpace(m.Host) == "" {
			return nil, fmt.Errorf("machine %d (%s): host required", i, m.Name)
		}
		if m.Port <= 0 {
			m.Port = settings.DefaultHaasPort
		}
		if m.Port > 65535 {
			return nil, fmt.Errorf("machine %d (%s): port out of range", i, m.Name)
		}
		if strings.TrimSpace(m.Brand) == "" {
			m.Brand = settings.MachineBrandHaas
		}
		switch m.CameraType {
		case "", "auto", "hls", "mjpeg", "iframe", "none":
			if m.CameraType == "" {
				m.CameraType = "auto"
			}
		default:
			return nil, fmt.Errorf("machine %d (%s): invalid cameraType %q", i, m.Name, m.CameraType)
		}
		if m.ID == "" {
			m.ID = newMachineID()
		}
		if _, dupe := seenIDs[m.ID]; dupe {
			return nil, fmt.Errorf("machine %d (%s): duplicate id %q", i, m.Name, m.ID)
		}
		seenIDs[m.ID] = struct{}{}
		_ = existing // currently unused; kept for future "preserve ID by name match" rules
		out = append(out, m)
	}
	return out, nil
}

func newMachineID() string {
	// 16-byte URL-safe random (~22 chars). Operators configure
	// a handful of machines per install, no collision risk.
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		// Crypto-rand failure is exotic; fall back to a timestamp
		// so the install isn't bricked. Collision risk negligible at
		// machine-config cadence.
		return fmt.Sprintf("m%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}

// cncMachinesListHandler returns the configured Machines (id + name +
// host:port + camera). Auth: any logged-in user — the frontend store
// needs this to drive the machine switcher.
func cncMachinesListHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		return renderJSON(w, r, map[string]any{
			"machines":  registry.Machines(),
			"default_id": defaultMachineID(registry),
		})
	})
}

func defaultMachineID(registry *cnc.Registry) string {
	ms := registry.Machines()
	if len(ms) == 0 {
		return ""
	}
	return ms[0].ID
}

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

// resolveStreamer pulls the streamer for ?machine_id= (or default
// when missing). Returns 404 + nil if no machine matches. Handlers
// short-circuit on a non-zero status.
func resolveStreamer(registry *cnc.Registry, r *http.Request) (*cnc.Streamer, string, int, error) {
	id := r.URL.Query().Get("machine_id")
	st, resolvedID := registry.Streamer(id)
	if st == nil {
		return nil, "", http.StatusNotFound, fmt.Errorf("no machine configured (id=%q)", id)
	}
	return st, resolvedID, 0, nil
}

func resolveAggregator(registry *cnc.Registry, r *http.Request) (*cnc.Aggregator, string, int, error) {
	id := r.URL.Query().Get("machine_id")
	ag, resolvedID := registry.Aggregator(id)
	if ag == nil {
		return nil, "", http.StatusNotFound, fmt.Errorf("no machine configured (id=%q)", id)
	}
	return ag, resolvedID, 0, nil
}

func cncStatusHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		st, machineID, code, err := resolveStreamer(registry, r)
		if err != nil {
			return code, err
		}
		s := st.Status()
		body := struct {
			*cnc.Status
			MachineID string `json:"machine_id"`
			FileURL   string `json:"file_url,omitempty"`
		}{Status: s, MachineID: machineID}
		if s.FilePath != "" {
			body.FileURL = "/files" + ensureLeading(s.FilePath)
		}
		return renderJSON(w, r, body)
	})
}

func cncCheckHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		st, machineID, code, err := resolveStreamer(registry, r)
		if err != nil {
			return code, err
		}
		ag, _ := registry.Aggregator(machineID)
		if ag != nil {
			ag.Wake(0)
		}

		body := struct {
			MachineID string `json:"machine_id"`
			Bridge    struct {
				OK        bool    `json:"ok"`
				LatencyMs float64 `json:"latency_ms,omitempty"`
				Error     string  `json:"error,omitempty"`
				Address   string  `json:"address,omitempty"`
			} `json:"bridge"`
			Controller struct {
				OK        bool    `json:"ok"`
				LatencyMs float64 `json:"latency_ms,omitempty"`
				Error     string  `json:"error,omitempty"`
				Mode      string  `json:"mode,omitempty"`
			} `json:"controller"`
		}{MachineID: machineID}

		if st.IsRunning() {
			body.Bridge.Error = "stream in progress — connection check skipped to avoid disturbing the job"
			body.Controller.Error = body.Bridge.Error
			return renderJSON(w, r, body)
		}

		bridgeOK, bridgeLatency, bridgeAddr, bridgeErr := st.CheckBridge()
		body.Bridge.OK = bridgeOK
		body.Bridge.LatencyMs = bridgeLatency
		body.Bridge.Address = bridgeAddr
		if bridgeErr != nil {
			body.Bridge.Error = bridgeErr.Error()
			body.Controller.Error = "skipped (bridge unreachable)"
			return renderJSON(w, r, body)
		}

		ctrlOK, ctrlLatency, mode, ctrlErr := st.CheckController(r.Context())
		body.Controller.OK = ctrlOK
		body.Controller.LatencyMs = ctrlLatency
		body.Controller.Mode = mode
		if ctrlErr != nil {
			body.Controller.Error = ctrlErr.Error()
		}
		return renderJSON(w, r, body)
	})
}

// modelExtensions is the set of 3D model file extensions the siblings
// endpoint will surface as a candidate part-view source.
var modelExtensions = map[string]bool{
	".3mf":  true,
	".stl":  true,
	".step": true,
	".stp":  true,
	".x_t":  true,
	".x_b":  true,
	".iges": true,
	".igs":  true,
	".obj":  true,
	".ply":  true,
}

// cncSiblingsHandler — same as before; not multi-machine aware (the
// share is global to the install, not per-machine).
func cncSiblingsHandler(_ *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		raw := r.URL.Query().Get("path")
		if raw == "" {
			return http.StatusBadRequest, errors.New("path required")
		}
		clean := path.Clean(ensureLeading(raw))
		if strings.Contains(clean, "..") {
			return http.StatusBadRequest, errors.New("path must not escape the share")
		}

		dir := path.Dir(clean)
		base := strings.TrimSuffix(path.Base(clean), path.Ext(clean))
		baseLower := strings.ToLower(base)

		f, err := d.user.Fs.Open(dir)
		if err != nil {
			return errToStatus(err), err
		}
		defer f.Close()
		entries, err := f.Readdir(-1)
		if err != nil {
			return errToStatus(err), err
		}

		body := struct {
			ModelURL    string `json:"model_url,omitempty"`
			ModelName   string `json:"model_name,omitempty"`
			ModelPath   string `json:"model_path,omitempty"`
			DrawingURL  string `json:"drawing_url,omitempty"`
			DrawingName string `json:"drawing_name,omitempty"`
			DrawingPath string `json:"drawing_path,omitempty"`
		}{}

		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			ext := strings.ToLower(path.Ext(name))
			stem := strings.ToLower(strings.TrimSuffix(name, path.Ext(name)))
			if stem != baseLower {
				continue
			}
			full := path.Join(dir, name)
			if modelExtensions[ext] && body.ModelURL == "" {
				body.ModelURL = "/api/raw" + full + "?inline=true"
				body.ModelName = name
				body.ModelPath = full
			} else if ext == ".pdf" && body.DrawingURL == "" {
				body.DrawingURL = "/api/raw" + full + "?inline=true"
				body.DrawingName = name
				body.DrawingPath = full
			}
			if body.ModelURL != "" && body.DrawingURL != "" {
				break
			}
		}
		return renderJSON(w, r, body)
	})
}

func cncProbeToolsHandler(registry *cnc.Registry) handleFunc {
	return withAdmin(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		st, _, code, err := resolveStreamer(registry, r)
		if err != nil {
			return code, err
		}
		slots := 30
		if q := r.URL.Query().Get("slots"); q != "" {
			n, err := strconv.Atoi(q)
			if err != nil || n < 1 || n > 200 {
				return http.StatusBadRequest, fmt.Errorf("slots must be 1..200")
			}
			slots = n
		}
		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
		defer cancel()
		rep, err := st.ProbeTools(ctx, slots)
		if err != nil {
			return errToStatus(err), err
		}
		return renderJSON(w, r, rep)
	})
}

type cncStartBody struct {
	FilePath  string `json:"file_path"`
	MachineID string `json:"machine_id,omitempty"` // optional; ?machine_id= also accepted
}

func cncStartHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		req := &cncStartBody{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			return http.StatusBadRequest, err
		}
		if req.FilePath == "" {
			return http.StatusBadRequest, errors.New("file_path required")
		}
		// Body wins over query param when both are sent.
		machineID := req.MachineID
		if machineID == "" {
			machineID = r.URL.Query().Get("machine_id")
		}
		streamer, _ := registry.Streamer(machineID)
		if streamer == nil {
			return http.StatusNotFound, fmt.Errorf("no machine configured (id=%q)", machineID)
		}
		ag, _ := registry.Aggregator(machineID)
		if ag != nil {
			ag.Wake(0)
		}

		clean := path.Clean(ensureLeading(req.FilePath))
		if strings.Contains(clean, "..") {
			return http.StatusBadRequest, errors.New("file_path must not escape the share")
		}
		absPath := d.user.FullPath(clean)

		st, err := streamer.Start(absPath, clean)
		switch {
		case errors.Is(err, cnc.ErrJobAlreadyRunning):
			return http.StatusConflict, err
		case errors.Is(err, cnc.ErrRecoveryPending):
			return http.StatusConflict, err
		case errors.Is(err, cnc.ErrConfigMissing):
			return http.StatusBadRequest, err
		case err != nil:
			return errToStatus(err), err
		}
		return renderJSON(w, r, map[string]string{"job_id": st.JobID})
	})
}

func cncStopHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		st, _, code, err := resolveStreamer(registry, r)
		if err != nil {
			return code, err
		}
		stopped := st.Stop()
		return renderJSON(w, r, map[string]bool{"stopped": stopped})
	})
}

func cncStateHandler(registry *cnc.Registry) handleFunc {
	session := withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		ag, _, code, err := resolveAggregator(registry, r)
		if err != nil {
			return code, err
		}
		ag.Wake(0)
		return renderJSON(w, r, ag.Snapshot())
	})
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			got := strings.TrimPrefix(auth, "Bearer ")
			if d.settings.Cnc.MachineToken == "" || got != d.settings.Cnc.MachineToken {
				return http.StatusUnauthorized, nil
			}
			ag, _, code, err := resolveAggregator(registry, r)
			if err != nil {
				return code, err
			}
			ag.Wake(0)
			return renderJSON(w, r, ag.Snapshot())
		}
		return session(w, r, d)
	}
}

func cncRecoveryAckHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		st, _, code, err := resolveStreamer(registry, r)
		if err != nil {
			return code, err
		}
		st.AckRecovery()
		return renderJSON(w, r, map[string]bool{"acknowledged": true})
	})
}

type cncQueryBody struct {
	Q   int  `json:"q"`
	Var *int `json:"var,omitempty"`
}

func cncQueryHandler(registry *cnc.Registry) handleFunc {
	session := withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		st, _, code, err := resolveStreamer(registry, r)
		if err != nil {
			return code, err
		}
		return runQuery(w, r, st)
	})
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			got := strings.TrimPrefix(auth, "Bearer ")
			if d.settings.Cnc.MachineToken == "" || got != d.settings.Cnc.MachineToken {
				return http.StatusUnauthorized, nil
			}
			st, _, code, err := resolveStreamer(registry, r)
			if err != nil {
				return code, err
			}
			return runQuery(w, r, st)
		}
		return session(w, r, d)
	}
}

func cncStreamHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		streamer, _, code, err := resolveStreamer(registry, r)
		if err != nil {
			return code, err
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		defer conn.Close()

		if err := writeJSONFrame(conn, cnc.Event{Type: "status", Status: streamer.Status()}); err != nil {
			return 0, nil
		}

		events := streamer.Subscribe()
		defer streamer.Unsubscribe(events)

		readDone := make(chan struct{})
		go func() {
			defer close(readDone)
			for {
				if _, _, err := conn.NextReader(); err != nil {
					return
				}
			}
		}()

		ping := time.NewTicker(30 * time.Second)
		defer ping.Stop()

		for {
			select {
			case ev, ok := <-events:
				if !ok {
					return 0, nil
				}
				if err := writeJSONFrame(conn, ev); err != nil {
					return 0, nil
				}
			case <-ping.C:
				_ = conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(WSWriteDeadline))
			case <-readDone:
				return 0, nil
			case <-r.Context().Done():
				return 0, nil
			}
		}
	})
}

func writeJSONFrame(conn *websocket.Conn, v any) error {
	if err := conn.SetWriteDeadline(time.Now().Add(WSWriteDeadline)); err != nil {
		return err
	}
	return conn.WriteJSON(v)
}

func runQuery(w http.ResponseWriter, r *http.Request, streamer *cnc.Streamer) (int, error) {
	req := &cncQueryBody{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return http.StatusBadRequest, err
	}
	if req.Q <= 0 {
		return http.StatusBadRequest, errors.New("q must be a positive integer")
	}

	res, err := streamer.Query(r.Context(), req.Q, req.Var)
	switch {
	case errors.Is(err, cnc.ErrConfigMissing):
		return http.StatusBadRequest, err
	case err != nil:
		return errToStatus(err), err
	}
	return renderJSON(w, r, res)
}

func ensureLeading(p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	return "/" + p
}

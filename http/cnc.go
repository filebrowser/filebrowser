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
	"errors"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/filebrowser/filebrowser/v2/cnc"
	"github.com/filebrowser/filebrowser/v2/settings"
)

// cncSettingsBody is the wire shape the Machine settings tab POSTs and
// reads. Keep it 1:1 with settings.Cnc so the JSON round-trips cleanly,
// EXCEPT MachineToken — minted server-side via the dedicated regenerate
// endpoint so a stray PUT can't blank it out.
type cncSettingsBody struct {
	HaasHost     string `json:"haasHost"`
	HaasPort     int    `json:"haasPort"`
	CameraURL    string `json:"cameraUrl"`
	MachineToken string `json:"machineToken,omitempty"` // GET only
}

func cncFromSettings(c settings.Cnc) cncSettingsBody {
	return cncSettingsBody{
		HaasHost:     c.HaasHost,
		HaasPort:     c.HaasPort,
		CameraURL:    c.CameraURL,
		MachineToken: c.MachineToken,
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
	// MachineToken intentionally not touched — see comment on the body type.

	err := d.store.Settings.Save(d.settings)
	return errToStatus(err), err
})

// cncRegenerateTokenHandler mints a fresh opaque secret for any
// external service (Home Assistant scripts, monitoring agents, custom
// dashboards) that needs to call /api/cnc/state or /api/cnc/qcode
// server-to-server without a filebrowser session. Old token is
// invalidated immediately. Admin-only — never exposed to non-admin
// sessions and never returned in a list endpoint.
//
// Originally minted for the haas-dashboard repo (now archived;
// filebrowser-NC subsumed its functionality). The mechanism stays
// because it's useful for any S2S consumer.
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

// cncStatusHandler is auth-required (any logged-in user) so the
// breadcrumb pill can poll it from every page. It does NOT require
// admin — operators need to see whether a job is running.
//
// Status comes from the live cnc.Streamer singleton; the FileURL
// field is composed here so the dashboard can deep-link straight to
// the file in filebrowser without re-deriving the path.
func cncStatusHandler(streamer *cnc.Streamer) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		st := streamer.Status()
		body := struct {
			*cnc.Status
			FileURL string `json:"file_url,omitempty"`
		}{Status: st}
		if st.FilePath != "" {
			body.FileURL = "/files" + ensureLeading(st.FilePath)
		}
		return renderJSON(w, r, body)
	})
}

// cncCheckHandler probes connectivity in two layers:
//
//   1. Bridge — TCP dial to the configured Haas host:port. Reports
//      whether the Waveshare RS-232↔TCP bridge is reachable on the
//      network (cabling / power / ip address sane).
//   2. Controller — sends a Q104 (mode) round-trip and validates the
//      response shape. If the bridge dials but the controller doesn't
//      answer or returns garbage, that points at Setting 143 / RS-232
//      cabling / pendant-off, not network. Auth: any logged-in user.
func cncCheckHandler(streamer *cnc.Streamer, agg *cnc.Aggregator) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		// Operator looked at the machine — extend the polling window
		// so the dashboard fills in normally afterwards.
		agg.Wake(0)
		body := struct {
			Bridge struct {
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
		}{}

		if streamer.IsRunning() {
			body.Bridge.Error = "stream in progress — connection check skipped to avoid disturbing the job"
			body.Controller.Error = body.Bridge.Error
			return renderJSON(w, r, body)
		}

		bridgeOK, bridgeLatency, bridgeAddr, bridgeErr := streamer.CheckBridge()
		body.Bridge.OK = bridgeOK
		body.Bridge.LatencyMs = bridgeLatency
		body.Bridge.Address = bridgeAddr
		if bridgeErr != nil {
			body.Bridge.Error = bridgeErr.Error()
			body.Controller.Error = "skipped (bridge unreachable)"
			return renderJSON(w, r, body)
		}

		// Bridge is up — exercise a Q104 to see if the controller is
		// actually answering. CheckController honors the same query
		// serialization as the rest of the streamer.
		ctrlOK, ctrlLatency, mode, ctrlErr := streamer.CheckController(r.Context())
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
// endpoint will surface as a candidate part-view source. Matches what
// Online3DViewer can render in-browser. Drawing PDFs are matched
// separately with a fixed `.pdf`.
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

// cncSiblingsHandler answers the question "given an NC file at this
// path, where do I find the matching 3D model and PDF drawing?".
// Match rule: same directory, same basename (case-insensitive), one
// of `modelExtensions` for the model and `.pdf` for the drawing.
//
// Auth: any logged-in user. The file system access is scoped to the
// caller's view (d.user.Fs), so users only see siblings inside their
// own scope. Returned URLs are share-relative — the frontend prefixes
// /api/raw or /files as needed.
func cncSiblingsHandler(streamer *cnc.Streamer) handleFunc {
	_ = streamer // reserved for future use (e.g. resolve via active job)
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
			// model_url / drawing_url are RAW endpoints — these are
			// fetched directly by the 3D viewer + opened in a new tab
			// for the PDF, so they must hit /api/raw, not the SPA route.
			// model_path / drawing_path are share-relative for any UI
			// that wants to deep-link back into the file browser.
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

// cncStartBody is the request body for POST /api/cnc/start. file_path is
// share-relative — the same shape filebrowser uses elsewhere.
type cncStartBody struct {
	FilePath string `json:"file_path"`
}

// cncStartHandler validates the file path stays under the user's scope
// then hands off to the streamer. Phase 2.1 doesn't yet enforce
// machine-token auth here — the streamer endpoints expect a logged-in
// filebrowser user; haas-dashboard's machine-token only matters for
// /api/cnc/qcode (Phase 2.2).
func cncStartHandler(streamer *cnc.Streamer, agg *cnc.Aggregator) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		// Operator launched a job — wake the aggregator so the post-job
		// dashboard refresh has a live polling window already in flight.
		agg.Wake(0)

		req := &cncStartBody{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			return http.StatusBadRequest, err
		}
		if req.FilePath == "" {
			return http.StatusBadRequest, errors.New("file_path required")
		}

		// Resolve under the user's scope via afero.BasePathFs so escape
		// attempts (../../etc/passwd) get clamped at the share root —
		// same gate the rest of the HTTP layer uses.
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

func cncStopHandler(streamer *cnc.Streamer) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		stopped := streamer.Stop()
		return renderJSON(w, r, map[string]bool{"stopped": stopped})
	})
}

// cncStateHandler exposes the curated Q-code snapshot maintained by
// the aggregator. Auth: same shape as /qcode — accept either a
// logged-in filebrowser session or a matching machine bearer token,
// so the dashboard's UI tiles AND any external service can read it.
func cncStateHandler(agg *cnc.Aggregator) handleFunc {
	session := withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		agg.Wake(0) // 0 = use the aggregator's default wake window
		return renderJSON(w, r, agg.Snapshot())
	})
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			got := strings.TrimPrefix(auth, "Bearer ")
			if d.settings.Cnc.MachineToken == "" || got != d.settings.Cnc.MachineToken {
				return http.StatusUnauthorized, nil
			}
			agg.Wake(0)
			return renderJSON(w, r, agg.Snapshot())
		}
		return session(w, r, d)
	}
}

// cncRecoveryAckHandler clears the pending-recovery flag (Z-15) so
// Start can succeed again. Modify-permission gate matches start/stop —
// recovering from a partial cut is an operator decision, not an admin
// settings one.
func cncRecoveryAckHandler(streamer *cnc.Streamer) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		streamer.AckRecovery()
		return renderJSON(w, r, map[string]bool{"acknowledged": true})
	})
}

// cncQueryBody mirrors haas-dashboard's POST /api/query so the dashboard
// can swap its base URL between direct-Waveshare and Pi-broker without
// other code changes (D-1 in the integration plan).
type cncQueryBody struct {
	Q   int  `json:"q"`
	Var *int `json:"var,omitempty"`
}

// cncQueryHandler accepts either a logged-in filebrowser session OR a
// matching Authorization: Bearer <MachineToken> header (the
// server-to-server path used by haas-dashboard). Session path defers
// auth to withUser; token path validates inline so the dashboard
// doesn't need a filebrowser user account.
func cncQueryHandler(streamer *cnc.Streamer) handleFunc {
	session := withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		return runQuery(w, r, streamer)
	})
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			got := strings.TrimPrefix(auth, "Bearer ")
			if d.settings.Cnc.MachineToken == "" || got != d.settings.Cnc.MachineToken {
				return http.StatusUnauthorized, nil
			}
			return runQuery(w, r, streamer)
		}
		return session(w, r, d)
	}
}

// cncStreamHandler upgrades to a WebSocket and pushes line/status events
// from the streamer until the client disconnects. Auth: any logged-in
// user (operators need to watch from any browser session).
//
// Send-only on the server side: we don't expect client messages, but we
// run a read loop anyway so the WS keep-alive ping/pong works and we
// notice client disconnects promptly.
func cncStreamHandler(streamer *cnc.Streamer) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		defer conn.Close()

		// Send the current status as the first frame so a freshly-
		// connecting client doesn't have to wait for the next event
		// to know whether a job is running.
		if err := writeJSONFrame(conn, cnc.Event{Type: "status", Status: streamer.Status()}); err != nil {
			return 0, nil
		}

		events := streamer.Subscribe()
		defer streamer.Unsubscribe(events)

		// Read pump (drops anything the client sends; closing the
		// connection on the client side surfaces here as ReadMessage
		// returning an error).
		readDone := make(chan struct{})
		go func() {
			defer close(readDone)
			for {
				if _, _, err := conn.NextReader(); err != nil {
					return
				}
			}
		}()

		// Heartbeat so a client behind a NAT/proxy that drops idle TCP
		// gets evicted and reconnects rather than hanging forever.
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

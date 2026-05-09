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
func cncStartHandler(streamer *cnc.Streamer) handleFunc {
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

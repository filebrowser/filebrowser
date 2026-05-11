package fbhttp

// /api/cnc/auto-send — opt-in pipeline that bundles preflight + start
// into a single round-trip. When the machine has AutoSendEnabled and
// preflight comes back all-green (no missing / empty / warn tools and
// no pending spindle swap), the file goes straight to MEM-tab Receive.
//
// CYCLE START is NOT triggered remotely — Haas doesn't expose that over
// RS-232 in a safe way. Operators still press the physical button. The
// "auto" here means "skip the wizard click-through", not "auto-cycle".
//
// Refusals are explicit and pre-emptive: the response body always
// includes the preflight summary + reason so the UI can fall back to
// the normal wizard flow without a second round-trip.

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/filebrowser/filebrowser/v2/cnc"
)

// cncAutoSendBody is the wire shape for POST /api/cnc/auto-send.
type cncAutoSendBody struct {
	FilePath  string `json:"file_path"`
	MachineID string `json:"machine_id,omitempty"`
	// Method mirrors /api/cnc/start. Optional; defaults to "mem"
	// because that's the only mode where auto-send is meaningful —
	// DNC drip already requires the operator to be at the controller.
	Method string `json:"method,omitempty"`
	// QueueID, when present, marks the queue row in-flight before
	// the streamer takes the job. Optional.
	QueueID string `json:"queue_id,omitempty"`
}

// autoSendResponse is the wire shape on success AND blocked. `started`
// distinguishes — when true, job_id holds the streamer job id; when
// false, blocked_reason explains why.
type autoSendResponse struct {
	Started       bool             `json:"started"`
	JobID         string           `json:"job_id,omitempty"`
	BlockedReason string           `json:"blocked_reason,omitempty"`
	Preflight     *cnc.Preflight   `json:"preflight,omitempty"`
}

// cncAutoSendHandler runs preflight, evaluates the auto-send gate,
// and either starts the job or returns the block reason. Returns 202
// Accepted when the job starts; 409 Conflict when gated; 400/404 on
// bad input. Always renders an autoSendResponse so the client can
// surface the preflight summary either way.
func cncAutoSendHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		req := &cncAutoSendBody{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			return http.StatusBadRequest, err
		}
		if req.FilePath == "" {
			return http.StatusBadRequest, errors.New("file_path required")
		}
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

		machineCfg := findMachine(d.settings, machineID)
		if machineCfg == nil {
			return http.StatusNotFound, fmt.Errorf("no machine configured (id=%q)", machineID)
		}
		if !machineCfg.AutoSendEnabled {
			return renderJSON(w, r, autoSendResponse{
				BlockedReason: "auto-send is not enabled for this machine (Settings → Machine)",
			})
		}

		clean := path.Clean(ensureLeading(req.FilePath))
		if strings.Contains(clean, "..") {
			return http.StatusBadRequest, errors.New("file_path must not escape the share")
		}
		absPath := d.user.FullPath(clean)

		// Build preflight against the latest tool-table dump. Auto-send
		// requires a fresh table — without one we can't classify tools,
		// so we refuse rather than send blind.
		var table *cnc.ToolTable
		dir := toolTableDirAbs(d, machineID)
		if latestPath, _ := newestJSONIn(dir); latestPath != "" {
			if buf, rerr := os.ReadFile(latestPath); rerr == nil {
				var t cnc.ToolTable
				if json.Unmarshal(buf, &t) == nil {
					table = &t
				}
			}
		}

		pf, perr := cnc.BuildPreflight(absPath, clean, machineID, table, currentSpindleTool(registry, machineID))
		if perr != nil {
			return http.StatusBadRequest, fmt.Errorf("preflight failed: %w", perr)
		}
		if reason := autoSendBlockReason(pf); reason != "" {
			return renderJSON(w, r, autoSendResponse{
				BlockedReason: reason,
				Preflight:     pf,
			})
		}

		method := cnc.NormalizeSendMethod(req.Method)
		if req.QueueID != "" {
			if qs := registry.Queues(); qs != nil {
				if _, qerr := qs.MarkSending(machineID, req.QueueID, string(method)); qerr == nil {
					streamer.EmitQueueSnapshot(qs.List(machineID))
				}
			}
		}
		st, err := streamer.Start(absPath, clean, method)
		if err != nil {
			if req.QueueID != "" {
				if qs := registry.Queues(); qs != nil {
					qs.ClearInFlight(machineID)
					streamer.EmitQueueSnapshot(qs.List(machineID))
				}
			}
			switch {
			case errors.Is(err, cnc.ErrJobAlreadyRunning),
				errors.Is(err, cnc.ErrRecoveryPending):
				return http.StatusConflict, err
			case errors.Is(err, cnc.ErrConfigMissing):
				return http.StatusBadRequest, err
			default:
				return errToStatus(err), err
			}
		}
		// 202 Accepted communicates "send pipeline initiated, you should
		// poll /api/cnc/status (or watch the WS) for line progress".
		w.WriteHeader(http.StatusAccepted)
		return renderJSON(w, r, autoSendResponse{
			Started:   true,
			JobID:     st.JobID,
			Preflight: pf,
		})
	})
}

// autoSendBlockReason returns a non-empty string when preflight surfaces
// any condition that should keep the operator in the manual wizard
// loop. Empty string = clear to send.
//
// Auto-send is intentionally strict: any warn / missing / empty /
// offline tool blocks. The wizard is still the right surface for
// resolving those — auto-send only short-circuits the unambiguous case.
func autoSendBlockReason(pf *cnc.Preflight) string {
	if pf == nil {
		return "preflight unavailable"
	}
	if pf.TableMissing {
		return "no tool-table read on file — read the table first"
	}
	if pf.Summary.Missing > 0 {
		return fmt.Sprintf("%d tool(s) missing from the table", pf.Summary.Missing)
	}
	if pf.Summary.Empty > 0 {
		return fmt.Sprintf("%d tool(s) report empty pocket", pf.Summary.Empty)
	}
	if pf.Summary.Offline > 0 {
		return fmt.Sprintf("%d tool(s) errored on the last read", pf.Summary.Offline)
	}
	if pf.Summary.Warn > 0 {
		return fmt.Sprintf("%d tool(s) flagged warn (diameter drift / cutter-comp)", pf.Summary.Warn)
	}
	if pf.SpindleSwap {
		return "spindle swap pending — confirm starting tool in the wizard"
	}
	return ""
}

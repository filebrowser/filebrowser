package fbhttp

// /api/cnc/attach — operator-marked "this filebrowser file is what the
// controller is actually running, even though we didn't send it via
// the bridge." Drives the /machine dashboard's follow-along when an
// NC program was loaded from SD card / Ethernet drop.
//
// Two endpoints:
//   POST /api/cnc/attach   { file_path, source? }
//   DELETE /api/cnc/attach
//
// Attachment is cleared automatically when a real streaming job starts
// (the job's own file is authoritative). Manual detach is for operator
// recovery — e.g. they switched programs on the controller.

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

type cncAttachBody struct {
	FilePath  string `json:"file_path"`
	MachineID string `json:"machine_id,omitempty"`
	Source    string `json:"source,omitempty"` // "manual" | "auto" — defaults to "manual"
}

// cncAttachHandler validates the file path against the user scope,
// marks it on the streamer, and emits a status broadcast so any open
// dashboards pick it up. Validates source — anything other than the
// known values normalizes to "manual" so a future client can't sneak
// in a junk tag.
func cncAttachHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		req := &cncAttachBody{}
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
		clean := path.Clean(ensureLeading(req.FilePath))
		if strings.Contains(clean, "..") {
			return http.StatusBadRequest, errors.New("file_path must not escape the share")
		}
		// Existence check — operator might paste a stale path. The
		// streamer doesn't otherwise care; verifying here gives a
		// cleaner error than a silent "attached but never loads."
		if _, err := os.Stat(d.user.FullPath(clean)); err != nil {
			return http.StatusNotFound, fmt.Errorf("file not in scope: %s", clean)
		}
		source := strings.ToLower(strings.TrimSpace(req.Source))
		if source != "manual" && source != "auto" {
			source = "manual"
		}
		if err := streamer.Attach(clean, source); err != nil {
			return http.StatusConflict, err
		}
		// Broadcast the new status snapshot so subscribed dashboards
		// update without polling.
		streamer.EmitStatus()
		return renderJSON(w, r, streamer.Status())
	})
}

func cncDetachHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		machineID := r.URL.Query().Get("machine_id")
		streamer, _ := registry.Streamer(machineID)
		if streamer == nil {
			return http.StatusNotFound, fmt.Errorf("no machine configured (id=%q)", machineID)
		}
		if streamer.Detach() {
			streamer.EmitStatus()
		}
		return renderJSON(w, r, streamer.Status())
	})
}

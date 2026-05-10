package fbhttp

// /api/cnc/queue/* — per-machine staging queue for NC sends.
//
// Shared across operators (one queue per machine, not per-user) so a
// second operator viewing /machine sees what the first staged.
// Persistence lives in cnc.QueueStore — see cnc/queue.go.
//
// Mutations broadcast a "queue" event on the per-machine WS stream so
// every connected client refreshes without polling.

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/filebrowser/filebrowser/v2/cnc"
)

// cncQueueListHandler — GET /api/cnc/queue?machine_id=
func cncQueueListHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		_, machineID, code, err := resolveStreamer(registry, r)
		if err != nil {
			return code, err
		}
		qs := registry.Queues()
		if qs == nil {
			return renderJSON(w, r, []cnc.QueueItem{})
		}
		return renderJSON(w, r, qs.List(machineID))
	})
}

type cncQueueAddBody struct {
	FilePath  string `json:"file_path"`
	MachineID string `json:"machine_id,omitempty"`
}

// cncQueueAddHandler — POST /api/cnc/queue. Adds a file to the queue.
// Resolves the file in the calling operator's scope to read its
// O-number + size at enqueue time. The on-disk queue stores only the
// scope-relative path; resolving again at send time works for any
// operator with read access to the share.
func cncQueueAddHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		req := &cncQueueAddBody{}
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
		absPath := d.user.FullPath(clean)
		qs := registry.Queues()
		if qs == nil {
			return http.StatusServiceUnavailable, errors.New("queue persistence unavailable")
		}
		// Resolve canonical machineID — Streamer() handles the
		// default-when-empty case, but qs.Add wants the resolved id.
		_, resolvedID, _, _ := resolveStreamer(registry, r)
		item, err := qs.Add(resolvedID, cnc.QueueItem{FilePath: clean}, absPath)
		if err != nil {
			return errToStatus(err), err
		}
		streamer.EmitQueueSnapshot(qs.List(resolvedID))
		return renderJSON(w, r, item)
	})
}

// cncQueueRemoveHandler — DELETE /api/cnc/queue/{id}
func cncQueueRemoveHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		_, machineID, code, err := resolveStreamer(registry, r)
		if err != nil {
			return code, err
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/cnc/queue/")
		if id == "" {
			return http.StatusBadRequest, errors.New("queue item id required")
		}
		qs := registry.Queues()
		if qs == nil {
			return http.StatusServiceUnavailable, errors.New("queue persistence unavailable")
		}
		if err := qs.Remove(machineID, id); err != nil {
			return errToStatus(err), err
		}
		if streamer, _ := registry.Streamer(machineID); streamer != nil {
			streamer.EmitQueueSnapshot(qs.List(machineID))
		}
		return renderJSON(w, r, map[string]bool{"removed": true})
	})
}

type cncQueueReorderBody struct {
	IDs []string `json:"ids"`
}

// cncQueueReorderHandler — PATCH /api/cnc/queue
func cncQueueReorderHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}
		_, machineID, code, err := resolveStreamer(registry, r)
		if err != nil {
			return code, err
		}
		req := &cncQueueReorderBody{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			return http.StatusBadRequest, err
		}
		qs := registry.Queues()
		if qs == nil {
			return http.StatusServiceUnavailable, errors.New("queue persistence unavailable")
		}
		if err := qs.Reorder(machineID, req.IDs); err != nil {
			return errToStatus(err), err
		}
		if streamer, _ := registry.Streamer(machineID); streamer != nil {
			streamer.EmitQueueSnapshot(qs.List(machineID))
		}
		return renderJSON(w, r, qs.List(machineID))
	})
}

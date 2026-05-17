package fbhttp

// /api/cnc/tool-library — operator uploads their Fusion 360 tool
// library export so the dashboard can enrich live tool-table rows
// with descriptions, vendor links, flute counts, and (eventually) a
// revolved-profile SVG. Admin-only on PUT/DELETE; user-readable on GET.

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/filebrowser/filebrowser/v2/cnc"
)

// uploadCap bounds the request body so a malicious operator can't
// dump a few GB into the config dir. A real Fusion export of ~50
// tools is ~150 KB; 4 MB is well past anything a shop would have.
const toolLibraryUploadCap = 4 * 1024 * 1024

func cncToolLibraryGetHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		store := registry.LibraryStore()
		if store == nil {
			return renderJSON(w, r, map[string]any{
				"data":            []any{},
				"loaded":          false,
				"assigned_slots":  []int{},
			})
		}
		lib := store.Library()
		if lib == nil {
			return renderJSON(w, r, map[string]any{
				"data":            []any{},
				"loaded":          false,
				"assigned_slots":  []int{},
			})
		}
		raw := lib.Raw()
		return renderJSON(w, r, map[string]any{
			"data":           raw.Data,
			"version":        raw.Version,
			"uploaded_at":    raw.UploadedAt,
			"loaded":         true,
			"assigned_slots": lib.AssignedSlots(),
		})
	})
}

// cncToolLibrarySlotHandler returns the entry for one pocket number.
// /api/cnc/tool-library/slot/{n}. 404 when no library or no slot.
func cncToolLibrarySlotHandler(registry *cnc.Registry) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		store := registry.LibraryStore()
		if store == nil {
			return http.StatusNotFound, nil
		}
		lib := store.Library()
		if lib == nil {
			return http.StatusNotFound, nil
		}
		// Path tail: /api/cnc/tool-library/slot/{n}
		path := r.URL.Path
		idx := strings.LastIndex(path, "/")
		if idx < 0 || idx+1 >= len(path) {
			return http.StatusBadRequest, errors.New("missing slot number")
		}
		n, err := strconv.Atoi(path[idx+1:])
		if err != nil {
			return http.StatusBadRequest, err
		}
		entry, ok := lib.Lookup(n)
		if !ok {
			return http.StatusNotFound, nil
		}
		return renderJSON(w, r, entry)
	})
}

// cncToolLibraryPutHandler replaces the stored library with the
// supplied JSON. Admin-only.
func cncToolLibraryPutHandler(registry *cnc.Registry) handleFunc {
	return withAdmin(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		store := registry.LibraryStore()
		if store == nil {
			return http.StatusServiceUnavailable, errors.New("tool-library store not initialised")
		}
		r.Body = http.MaxBytesReader(w, r.Body, toolLibraryUploadCap)
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			return http.StatusBadRequest, err
		}
		lib, err := store.Replace(buf)
		if err != nil {
			return http.StatusBadRequest, err
		}
		raw := lib.Raw()
		return renderJSON(w, r, map[string]any{
			"loaded":         true,
			"uploaded_at":    raw.UploadedAt,
			"assigned_slots": lib.AssignedSlots(),
			"count":          len(raw.Data),
		})
	})
}

// cncToolLibraryDeleteHandler clears the stored library. Admin-only.
func cncToolLibraryDeleteHandler(registry *cnc.Registry) handleFunc {
	return withAdmin(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		store := registry.LibraryStore()
		if store == nil {
			return http.StatusServiceUnavailable, errors.New("tool-library store not initialised")
		}
		if err := store.Clear(); err != nil {
			return http.StatusInternalServerError, err
		}
		return renderJSON(w, r, map[string]bool{"cleared": true})
	})
}

package fbhttp

// /api/cnc/codes/* — Haas alarm/setting/parameter code lookup.
// Lets the UI translate a bare number like "Setting 414" into a
// human-readable explanation. Read-only; no registry interaction.

import (
	"net/http"
	"strconv"

	"github.com/filebrowser/filebrowser/v2/cnc"
)

// GET /api/cnc/codes/lookup?kind=setting&number=414
//
// Returns the curated entry if known, ok=false otherwise. Falls back to
// kind=setting if `kind` is unrecognized (operators reach this from the
// "Setting 414" tray more than the alarm tray).
var cncCodesLookupHandler = withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
	kind := cnc.NormalizeKind(r.URL.Query().Get("kind"))
	nStr := r.URL.Query().Get("number")
	n, err := strconv.Atoi(nStr)
	if err != nil {
		return http.StatusBadRequest, err
	}
	entry, ok := cnc.LookupCode(kind, n)
	return renderJSON(w, r, map[string]any{
		"ok":    ok,
		"kind":  string(kind),
		"entry": entry,
	})
})

// GET /api/cnc/codes/search?q=probe&kind=setting&limit=20
//
// Free-text scan over title + summary + hint + category. Both kind and
// q are optional. The UI uses this for the "what was that alarm again?"
// lookup field.
var cncCodesSearchHandler = withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
	q := r.URL.Query().Get("q")
	kindRaw := r.URL.Query().Get("kind")
	var kind cnc.CodeKind
	if kindRaw != "" {
		kind = cnc.NormalizeKind(kindRaw)
	}
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	results := cnc.SearchCodes(kind, q, limit)
	return renderJSON(w, r, map[string]any{
		"count":   len(results),
		"results": results,
	})
})

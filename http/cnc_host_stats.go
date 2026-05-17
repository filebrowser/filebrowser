package fbhttp

// GET /api/cnc/host-stats — small Pi-side health snapshot for the
// global header pill (Pi 4 SoC temp, load avg, mem/disk %, uptime).
// Cheap: ~5 small file reads + one statfs per call. User-authed so
// any operator viewing the dashboard sees it; admin not required.

import (
	"net/http"

	"github.com/filebrowser/filebrowser/v2/cnc"
)

var cncHostStatsHandler = withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
	return renderJSON(w, r, cnc.ReadHostStats())
})

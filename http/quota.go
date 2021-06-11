package http

import (
	"encoding/json"
	"net/http"
	"os"
)

type quotaData struct {
	InodeLimit uint64 `json:"inodeLimit"`
	InodeQuota uint64 `json:"inodeQuota"`
	InodeUsage uint64 `json:"inodeUsage"`
	SpaceLimit uint64 `json:"spaceLimit"`
	SpaceQuota uint64 `json:"spaceQuota"`
	SpaceUsage uint64 `json:"spaceUsage"`
}

var quotaGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	content, err := os.ReadFile(d.user.QuotaFile)
	if err != nil {
		return errToStatus(err), err
	}

	data := quotaData{}
	err = json.Unmarshal(content, &data)
	if err != nil {
		return errToStatus(err), err
	}

	res := map[string]map[string]uint64{
		"inodes": {
			"quota": data.InodeQuota,
			"usage": data.InodeUsage,
		},
		"space": {
			"quota": data.SpaceQuota,
			"usage": data.SpaceUsage,
		},
	}

	return renderJSON(w, r, res)
})

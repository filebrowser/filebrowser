package http

import (
	"net/http"

	"github.com/filebrowser/filebrowser/v2/users"
)

type contextInfo struct {
	Bookmarks []users.Bookmark `json:"bookmarks"`
}

var contextGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	context := &contextInfo{
		Bookmarks: d.user.Bookmarks,
	}

	return renderJSON(w, r, context)
})

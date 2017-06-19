package http

import (
	"net/http"
	"strings"

	fm "github.com/hacdias/filemanager"
	"github.com/hacdias/filemanager/page"
)

// serveSingle serves a single file in an editor (if it is editable), shows the
// plain file, or downloads it if it can't be shown.
func serveSingle(w http.ResponseWriter, r *http.Request, c *fm.Config, u *fm.User, i *fm.FileInfo) (int, error) {
	var err error

	if err = i.RetrieveFileType(); err != nil {
		return errorToHTTPCode(err, true), err
	}

	p := &page.Page{
		Info: &page.Info{
			Name:   i.Name,
			Path:   i.VirtualPath,
			IsDir:  false,
			Data:   i,
			User:   u,
			Config: c,
		},
	}

	// If the request accepts JSON, we send the file information.
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		return p.PrintAsJSON(w)
	}

	if i.Type == "text" {
		if err = i.Read(); err != nil {
			return errorToHTTPCode(err, true), err
		}
	}

	if i.CanBeEdited() && u.AllowEdit {
		p.Data, err = getEditor(r, i)
		p.Editor = true
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return p.PrintAsHTML(w, "frontmatter", "editor")
	}

	return p.PrintAsHTML(w, "single")
}

package filemanager

import (
	"net/http"
	"strings"
)

// serveSingle serves a single file in an editor (if it is editable), shows the
// plain file, or downloads it if it can't be shown.
func (c *Config) serveSingle(w http.ResponseWriter, r *http.Request, u *User, i *file) (int, error) {
	var err error

	if err = i.RetrieveFileType(); err != nil {
		return errorToHTTPCode(err, true), err
	}

	p := &page{
		Info: &pageInfo{
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
		return p.PrintJSON(w)
	}

	if i.Type == "text" {
		if err = i.Read(); err != nil {
			return errorToHTTPCode(err, true), err
		}
	}

	if i.CanBeEdited() && u.AllowEdit {
		p.Info.Data, err = newEditor(r, i)
		p.Info.Editor = true
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return p.PrintHTML(w, "frontmatter", "editor")
	}

	return p.PrintHTML(w, "single")
}

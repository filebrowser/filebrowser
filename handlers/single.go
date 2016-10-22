package handlers

import (
	"net/http"

	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/file"
	"github.com/hacdias/caddy-filemanager/page"
	"github.com/hacdias/caddy-filemanager/utils"
)

// ServeSingle serves a single file in an editor (if it is editable), shows the
// plain file, or downloads it if it can't be shown.
func ServeSingle(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.User, i *file.Info) (int, error) {
	err := i.Read()
	if err != nil {
		return utils.ErrorToHTTPCode(err, true), err
	}

	p := &page.Page{
		Info: &page.Info{
			Name:   i.Name(),
			Path:   i.VirtualPath,
			IsDir:  false,
			Data:   i,
			User:   u,
			Config: c,
		},
	}

	if i.CanBeEdited() && u.AllowEdit {
		p.Data, err = i.GetEditor()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return p.PrintAsHTML(w, "frontmatter", "editor")
	}

	return p.PrintAsHTML(w, "single")
}

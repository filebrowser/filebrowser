package filemanager

import (
	"net/http"
	"strings"
)

// serveSingle serves a single file in an editor (if it is editable), shows the
// plain file, or downloads it if it can't be shown.
func serveSingle(w http.ResponseWriter, r *http.Request, c *FileManager, u *user, i *fileInfo) (int, error) {
	var err error

	if err = i.RetrieveFileType(); err != nil {
		return errorToHTTP(err, true), err
	}

	p := &page{
		Name:      i.Name,
		Path:      i.VirtualPath,
		IsDir:     false,
		Data:      i,
		User:      u,
		PrefixURL: c.PrefixURL,
		BaseURL:   c.AbsoluteURL(),
		WebDavURL: c.AbsoluteWebDavURL(),
	}

	// If the request accepts JSON, we send the file information.
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		return p.PrintAsJSON(w)
	}

	if i.Type == "text" {
		if err = i.Read(); err != nil {
			return errorToHTTP(err, true), err
		}
	}

	if i.CanBeEdited() && u.AllowEdit {
		p.Data, err = GetEditor(r, i)
		p.Editor = true
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return p.PrintAsHTML(w, c.Assets.Templates, "frontmatter", "editor")
	}

	return p.PrintAsHTML(w, c.Assets.Templates, "single")
}

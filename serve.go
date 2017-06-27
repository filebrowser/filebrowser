package filemanager

import (
	"net/http"
	"strconv"
	"strings"
)

func serveDefault(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var err error

	c.pg = &page{
		Name:      c.fi.Name,
		Path:      c.fi.VirtualPath,
		User:      c.us,
		BaseURL:   c.fm.RootURL(),
		WebDavURL: c.fm.WebDavURL(),
	}

	// If it is a dir, go and serve the listing.
	if c.fi.IsDir {
		return serveListing(c, w, r)
	}

	// Tries to get the file type.
	if err = c.fi.RetrieveFileType(); err != nil {
		return errorToHTTP(err, true), err
	}

	// If the request accepts JSON, we send the file information.
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		c.pg.Data = c.fi
		return c.pg.PrintJSON(w)
	}

	// If it is a text file, reads its content.
	if c.fi.Type == "text" {
		if err = c.fi.Read(); err != nil {
			return errorToHTTP(err, true), err
		}
	}

	// If it can't be edited or the user isn't allowed to,
	// serve it as a listing, with a preview of the file.
	if !c.fi.CanBeEdited() || !c.us.AllowEdit {
		return serveListing(c, w, r)
	}

	// Otherwise, we just bring the editor in!
	c.pg.IsEditor = true
	c.pg.Data, err = getEditor(r, c.fi)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return c.pg.PrintHTML(w, c.fm.assets.templates, "frontmatter", "editor")
}

// serveListing presents the user with a listage of a directory folder.
func serveListing(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var (
		err     error
		listing *listing
	)

	c.pg.minimal = (r.Header.Get("Minimal") == "true")
	c.pg.IsEditor = false

	if c.fi.IsDir {
		// In this case, we are just showing a listing.
		listing, err = getListing(c.us, c.fi.VirtualPath, c.fm.RootURL()+r.URL.Path)
	} else {
		// But in this case we are showing a listing alongside with a preview!
		vpath := strings.TrimSuffix(c.fi.VirtualPath, c.fi.Name)
		url := strings.TrimSuffix(r.URL.Path, c.fi.Name)

		listing, err = getListing(c.us, vpath, url)
	}

	if err != nil {
		return errorToHTTP(err, true), err
	}

	cookieScope := c.fm.RootURL()
	if cookieScope == "" {
		cookieScope = "/"
	}

	// Copy the query values into the Listing struct
	var limit int
	listing.Sort, listing.Order, limit, err = handleSortOrder(w, r, cookieScope)
	if err != nil {
		return http.StatusBadRequest, err
	}

	listing.ApplySort()

	if limit > 0 && limit <= len(listing.Items) {
		listing.Items = listing.Items[:limit]
		listing.ItemsLimitedTo = limit
	}

	// If it's a JSON request, print only the items... in JSON! (such a surprise -_-)
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		c.pg.Data = listing.Items
		return c.pg.PrintJSON(w)
	}

	if !c.fi.IsDir {
		listing.Preview = true
		listing.PreviewItem = c.fi
	}

	listing.Display = displayMode(w, r, cookieScope)
	c.pg.Data = listing
	return c.pg.PrintHTML(w, c.fm.assets.templates, "listing")
}

// displayMode obtaisn the display mode from URL, or from the
// cookie.
func displayMode(w http.ResponseWriter, r *http.Request, scope string) string {
	displayMode := r.URL.Query().Get("display")

	if displayMode == "" {
		if displayCookie, err := r.Cookie("display"); err == nil {
			displayMode = displayCookie.Value
		}
	}

	if displayMode == "" || (displayMode != "mosaic" && displayMode != "list") {
		displayMode = "mosaic"
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "display",
		Value:  displayMode,
		Path:   scope,
		Secure: r.TLS != nil,
	})

	return displayMode
}

// handleSortOrder gets and stores for a Listing the 'sort' and 'order',
// and reads 'limit' if given. The latter is 0 if not given. Sets cookies.
func handleSortOrder(w http.ResponseWriter, r *http.Request, scope string) (sort string, order string, limit int, err error) {
	sort = r.URL.Query().Get("sort")
	order = r.URL.Query().Get("order")
	limitQuery := r.URL.Query().Get("limit")

	// If the query 'sort' or 'order' is empty, use defaults or any values
	// previously saved in Cookies.
	switch sort {
	case "":
		sort = "name"
		if sortCookie, sortErr := r.Cookie("sort"); sortErr == nil {
			sort = sortCookie.Value
		}
	case "name", "size", "type":
		http.SetCookie(w, &http.Cookie{
			Name:   "sort",
			Value:  sort,
			Path:   scope,
			Secure: r.TLS != nil,
		})
	}

	switch order {
	case "":
		order = "asc"
		if orderCookie, orderErr := r.Cookie("order"); orderErr == nil {
			order = orderCookie.Value
		}
	case "asc", "desc":
		http.SetCookie(w, &http.Cookie{
			Name:   "order",
			Value:  order,
			Path:   scope,
			Secure: r.TLS != nil,
		})
	}

	if limitQuery != "" {
		limit, err = strconv.Atoi(limitQuery)
		// If the 'limit' query can't be interpreted as a number, return err.
		if err != nil {
			return
		}
	}

	return
}

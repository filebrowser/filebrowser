package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/file"
	"github.com/hacdias/caddy-filemanager/page"
	"github.com/hacdias/caddy-filemanager/utils/errors"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// ServeListing presents the user with a listage of a directory folder.
func ServeListing(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.User, i *file.Info) (int, error) {
	var err error

	// Loads the content of the directory
	listing, err := file.GetListing(u, i.VirtualPath, c.PrefixURL+r.URL.Path)
	if err != nil {
		return errors.ErrorToHTTPCode(err, true), err
	}

	listing.Context = httpserver.Context{
		Root: http.Dir(u.Scope),
		Req:  r,
		URL:  r.URL,
	}

	cookieScope := c.BaseURL
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

	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		marsh, err := json.Marshal(listing.Items)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if _, err := w.Write(marsh); err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}

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
		Path:   cookieScope,
		Secure: r.TLS != nil,
	})

	page := &page.Page{
		Minimal: r.Header.Get("Minimal") == "true",
		Info: &page.Info{
			Name:    listing.Name,
			Path:    i.VirtualPath,
			IsDir:   true,
			User:    u,
			Config:  c,
			Display: displayMode,
			Data:    listing,
		},
	}

	return page.PrintAsHTML(w, "listing")
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

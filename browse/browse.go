package browse

import (
	"net/http"

	"github.com/hacdias/caddy-hugo/page"
)

// Execute sth
func Execute(w http.ResponseWriter, r *http.Request) (int, error) {
	page := new(page.Page)
	page.Title = "Browse"
	return page.Render(w, r, "browse")
}

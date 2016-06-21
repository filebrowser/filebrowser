// Package hugo makes the bridge between the static website generator Hugo
// and the webserver Caddy, also providing an administrative user interface.
package hugo

import (
	"net/http"

	"github.com/hacdias/caddy-filemanager"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// Hugo contais the next middleware to be run and the configuration
// of the current one.
type Hugo struct {
	FileManager *filemanager.FileManager
	Next        httpserver.Handler
	Config      *Config
}

// ServeHTTP is the main function of the whole plugin that routes every single
// request to its function.
func (h Hugo) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {

	return h.Next.ServeHTTP(w, r)
}

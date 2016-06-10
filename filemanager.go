package filemanager

import (
	"net/http"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

type FileManager struct {
	Next httpserver.Handler
}

func (f FileManager) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {

	return f.Next.ServeHTTP(w, r)
}

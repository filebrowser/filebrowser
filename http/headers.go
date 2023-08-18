//go:build !dev
// +build !dev

package http

// global headers to append to every response
var global_headers = map[string]string{
	"Cache-Control": "no-cache, no-store, must-revalidate",
}

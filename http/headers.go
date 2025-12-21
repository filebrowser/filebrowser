//go:build !dev

package fbhttp

// global headers to append to every response
var globalHeaders = map[string]string{
	"Cache-Control": "no-cache, no-store, must-revalidate",
}

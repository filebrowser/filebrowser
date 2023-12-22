//go:build dev
// +build dev

package http

// global headers to append to every response
// cross-origin headers are necessary to be able to
// access them from a different URL during development
var globalHeaders = map[string]string{
	"Cache-Control":                    "no-cache, no-store, must-revalidate",
	"Access-Control-Allow-Origin":      "*",
	"Access-Control-Allow-Headers":     "*",
	"Access-Control-Allow-Methods":     "*",
	"Access-Control-Allow-Credentials": "true",
}

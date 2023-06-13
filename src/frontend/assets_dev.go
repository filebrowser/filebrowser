//go:build dev
// +build dev

package frontend

import (
	"io/fs"
	"os"
)

var assets fs.FS = os.DirFS("frontend")

func Assets() fs.FS {
	return assets
}

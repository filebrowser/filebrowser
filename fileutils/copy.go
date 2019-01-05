package fileutils

import (
	"os"
	"path"

	"github.com/spf13/afero"
)

// Copy copies a file or folder from one place to another.
func Copy(fs afero.Fs, src, dst string) error {
	if src = path.Clean("/" + src); src == "" {
		return os.ErrNotExist
	}

	if dst = path.Clean("/" + dst); dst == "" {
		return os.ErrNotExist
	}

	if src == "/" || dst == "/" {
		// Prohibit copying from or to the virtual root directory.
		return os.ErrInvalid
	}

	if dst == src {
		return os.ErrInvalid
	}

	info, err := fs.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return CopyDir(fs, src, dst)
	}

	return CopyFile(fs, src, dst)
}

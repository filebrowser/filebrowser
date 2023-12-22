package fileutils

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// Copy copies a file or folder from one place to another.
func Copy(fs afero.Fs, src, dst, scope string) error {
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

	//nolint:exhaustive
	switch info.Mode() & os.ModeType {
	case os.ModeDir:
		return CopyDir(fs, src, dst, scope)
	case os.ModeSymlink:
		return CopySymLink(fs, src, dst, scope)
	default:
		return CopyFile(fs, src, dst)
	}
}

func CopySymLink(fs afero.Fs, source, dest, scope string) error {
	if reader, ok := fs.(afero.LinkReader); ok {
		link, err := reader.ReadlinkIfPossible(source)
		if err != nil {
			return err
		}

		if filepath.IsAbs(link) {
			link = strings.TrimPrefix(link, scope)
			link = filepath.Join(string(os.PathSeparator), link)
		} else {
			link = filepath.Clean(filepath.Join(filepath.Dir(source), link))
		}

		if linker, ok := fs.(afero.Linker); ok {
			return linker.SymlinkIfPossible(link, dest)
		}
		return nil
	}
	return nil
}

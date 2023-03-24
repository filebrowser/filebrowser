package fileutils

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// CopyDir copies a directory from source to dest and all
// of its sub-directories. It doesn't stop if it finds an error
// during the copy. Returns an error if any.
func CopyDir(fs afero.Fs, source, dest, scope string) error {
	// Get properties of source.
	srcinfo, err := fs.Stat(source)
	if err != nil {
		return err
	}

	// Create the destination directory.
	err = fs.MkdirAll(dest, srcinfo.Mode())
	if err != nil {
		return err
	}

	dir, _ := fs.Open(source)
	obs, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	var errs []error

	for _, obj := range obs {
		fsource := source + "/" + obj.Name()
		fdest := dest + "/" + obj.Name()

		switch obj.Mode() & os.ModeType {
		case os.ModeDir:
			// Create sub-directories, recursively.
			if err := CopyDir(fs, fsource, fdest, scope); err != nil {
				errs = append(errs, err)
			}
		case os.ModeSymlink:
			if err := CopySymLink(fs, fsource, fdest, scope); err != nil {
				return err
			}
		default:
			// Perform the file copy.
			if err := CopyFile(fs, fsource, fdest); err != nil {
				errs = append(errs, err)
			}
		}
	}

	var errString string
	for _, err := range errs {
		errString += err.Error() + "\n"
	}

	if errString != "" {
		return errors.New(errString)
	}

	return nil
}

func DiskUsage(fs afero.Fs, path string) (size, inodes int64, err error) {
	lst, ok := fs.(afero.Lstater)
	if !ok {
		return 0, 0, err
	}

	info, _, err := lst.LstatIfPossible(path)
	if err != nil {
		return 0, 0, err
	}

	size = info.Size()
	inodes = int64(1)

	// don't follow symlinks
	if !info.IsDir() {
		return size, inodes, err
	}

	afs := &afero.Afero{Fs: fs}
	dir, err := afs.ReadDir(path)
	if err != nil {
		return size, inodes, err
	}

	for _, fi := range dir {
		s, i, e := DiskUsage(fs, filepath.Join(path, fi.Name()))
		if e == nil {
			size += s
			inodes += i
		}
	}

	return size, inodes, err
}

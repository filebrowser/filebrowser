package fileutils

import (
	"errors"
	"path/filepath"

	"github.com/spf13/afero"
)

// CopyDir copies a directory from source to dest and all
// of its sub-directories. It doesn't stop if it finds an error
// during the copy. Returns an error if any.
func CopyDir(fs afero.Fs, source, dest string) error {
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

		if obj.IsDir() {
			// Create sub-directories, recursively.
			err = CopyDir(fs, fsource, fdest)
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			// Perform the file copy.
			err = CopyFile(fs, fsource, fdest)
			if err != nil {
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

func DiskUsage(fs afero.Fs, path string, maxDepth int) (size, inodes int64, err error) {
	info, err := fs.Stat(path)
	if err != nil {
		return 0, 0, err
	}

	size = info.Size()
	inodes = int64(1)

	if !info.IsDir() {
		return size, inodes, err
	}

	if maxDepth < 1 {
		return size, inodes, err
	}

	dir, err := fs.Open(path)
	if err != nil {
		return size, inodes, err
	}
	defer dir.Close()

	fis, err := dir.Readdir(-1)
	if err != nil {
		return size, inodes, err
	}

	for _, fi := range fis {
		if fi.Name() == "." || fi.Name() == ".." {
			continue
		}
		s, i, e := DiskUsage(fs, filepath.Join(path, fi.Name()), maxDepth-1)
		if e != nil {
			return size, inodes, e
		}
		size += s
		inodes += i
	}

	return size, inodes, err
}

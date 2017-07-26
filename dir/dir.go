// Package dir implements a FileSystem interface using the native
// file system restricted to a specific directory tree. Originally from
// https://github.com/golang/net/blob/master/webdav/file.go#L68
package dir

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// A Dir uses the native file system restricted to a specific directory tree.
//
// While the FileSystem.OpenFile method takes '/'-separated paths, a Dir's
// string value is a filename on the native file system, not a URL, so it is
// separated by filepath.Separator, which isn't necessarily '/'.
//
// An empty Dir is treated as ".".
type Dir string

func (d Dir) resolve(name string) string {
	// This implementation is based on Dir.Open's code in the standard net/http package.
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return ""
	}

	dir := string(d)
	if dir == "" {
		dir = "."
	}

	return filepath.Join(dir, filepath.FromSlash(SlashClean(name)))
}

// Mkdir implements os.Mkdir in this directory context.
func (d Dir) Mkdir(name string, perm os.FileMode) error {
	if name = d.resolve(name); name == "" {
		return os.ErrNotExist
	}
	return os.Mkdir(name, perm)
}

// OpenFile implements os.OpenFile in this directory context.
func (d Dir) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	if name = d.resolve(name); name == "" {
		return nil, os.ErrNotExist
	}
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// RemoveAll implements os.RemoveAll in this directory context.
func (d Dir) RemoveAll(name string) error {
	if name = d.resolve(name); name == "" {
		return os.ErrNotExist
	}

	if name == filepath.Clean(string(d)) {
		// Prohibit removing the virtual root directory.
		return os.ErrInvalid
	}
	return os.RemoveAll(name)
}

// Rename implements os.Rename in this directory context.
func (d Dir) Rename(oldName, newName string) error {
	if oldName = d.resolve(oldName); oldName == "" {
		return os.ErrNotExist
	}
	if newName = d.resolve(newName); newName == "" {
		return os.ErrNotExist
	}
	if root := filepath.Clean(string(d)); root == oldName || root == newName {
		// Prohibit renaming from or to the virtual root directory.
		return os.ErrInvalid
	}
	return os.Rename(oldName, newName)
}

// Stat implements os.Stat in this directory context.
func (d Dir) Stat(name string) (os.FileInfo, error) {
	if name = d.resolve(name); name == "" {
		return nil, os.ErrNotExist
	}

	return os.Stat(name)
}

// Copy copies a file or directory from src to dst. If it is
// a directory, all of the files and sub-directories will be copied.
func (d Dir) Copy(src, dst string) error {
	if src = d.resolve(src); src == "" {
		return os.ErrNotExist
	}

	if dst = d.resolve(dst); dst == "" {
		return os.ErrNotExist
	}

	if root := filepath.Clean(string(d)); root == src || root == dst {
		// Prohibit copying from or to the virtual root directory.
		return os.ErrInvalid
	}

	if dst == src {
		return os.ErrInvalid
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return CopyDir(src, dst)
	}

	return CopyFile(src, dst)
}

// SlashClean is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func SlashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}

// CopyFile copies a file from source to dest and returns
// an error if any.
func CopyFile(source string, dest string) error {
	// Open the source file.
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	// Makes the directory needed to create the dst
	// file.
	err = os.MkdirAll(filepath.Dir(dest), 0666)
	if err != nil {
		return err
	}

	// Create the destination file.
	dst, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy the contents of the file.
	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	// Copy the mode if the user can't
	// open the file.
	info, err := os.Stat(source)
	if err != nil {
		err = os.Chmod(dest, info.Mode())
		if err != nil {
			return err
		}
	}

	return nil
}

// CopyDir copies a directory from source to dest and all
// of its sub-directories. It doesn't stop if it finds an error
// during the copy. Returns an error if any.
func CopyDir(source string, dest string) error {
	// Get properties of source.
	srcinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// Create the destination directory.
	err = os.MkdirAll(dest, srcinfo.Mode())
	if err != nil {
		return err
	}

	dir, _ := os.Open(source)
	obs, err := dir.Readdir(-1)

	var errs []error

	for _, obj := range obs {
		fsource := source + "/" + obj.Name()
		fdest := dest + "/" + obj.Name()

		if obj.IsDir() {
			// Create sub-directories, recursively.
			err = CopyDir(fsource, fdest)
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			// Perform the file copy.
			err = CopyFile(fsource, fdest)
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

package fileutils

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// CopyFile copies a file from source to dest and returns
// an error if any.
func CopyFile(fs afero.Fs, source, dest string) error {
	// Open the source file.
	src, err := fs.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	// Makes the directory needed to create the dst
	// file.
	err = fs.MkdirAll(filepath.Dir(dest), 0666)
	if err != nil {
		return err
	}

	// Create the destination file.
	dst, err := fs.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
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
	info, err := fs.Stat(source)
	if err != nil {
		err = fs.Chmod(dest, info.Mode())
		if err != nil {
			return err
		}
	}

	return nil
}

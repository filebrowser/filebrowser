package files

import (
	"io"
	"os"
)

// CopyFile is used to copy a file
func CopyFile(old, new string) error {
	// Open the file and create a new one
	r, err := os.Open(old)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(new)
	if err != nil {
		return err
	}
	defer w.Close()

	// Copy the content
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}

	return nil
}

package types

import (
	"os"
)

func checkFS(path string) error {
	info, err := os.Stat(path)

	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		err = os.MkdirAll(path, 0666)
		if err != nil {
			return err
		}

		return nil
	}

	if !info.IsDir() {
		return ErrIsNotDirectory
	}

	return nil
}

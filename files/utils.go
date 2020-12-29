package files

import (
	"os"
)

func IsNamedPipe(mode os.FileMode) bool {
	return mode&os.ModeNamedPipe != 0
}

func IsSymlink(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0
}

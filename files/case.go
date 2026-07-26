package files

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/afero"
)

// CaseInsensitive reports whether the filesystem backing root treats file names
// case-insensitively, as NTFS, APFS, HFS+, exFAT and CIFS do. It creates a
// probe file and looks it up under a different case, mirroring how git detects
// core.ignoreCase.
//
// It probes rather than inferring from runtime.GOOS because neither direction
// holds: macOS can be formatted case-sensitively, and a Linux host commonly
// serves a case-insensitive mount. When the root cannot be written to, it falls
// back to the host's usual default rather than assuming case-sensitivity, since
// a read-only root is still exposed to the disclosure this guards against.
func CaseInsensitive(fs afero.Fs, root string) bool {
	probe, err := afero.TempFile(fs, root, "fb-case-probe-")
	if err != nil {
		return runtime.GOOS == "windows" || runtime.GOOS == "darwin"
	}

	name := probe.Name()
	probe.Close()
	defer fs.Remove(name) //nolint:errcheck

	dir, base := filepath.Split(name)
	_, err = fs.Stat(filepath.Join(dir, strings.ToUpper(base)))
	return err == nil
}

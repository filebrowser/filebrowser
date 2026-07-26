package files

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
)

// MemMapFs keys its entries by exact name, so it is case-sensitive.
func TestCaseInsensitiveReportsCaseSensitiveFs(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := fs.MkdirAll("/root", 0o755); err != nil {
		t.Fatal(err)
	}

	if CaseInsensitive(fs, "/root") {
		t.Error("CaseInsensitive() = true for a case-sensitive filesystem; want false")
	}
	assertNoProbeLeft(t, fs, "/root")
}

// Whatever the verdict, the probe file must not be left behind in the user's
// data directory.
func TestCaseInsensitiveRemovesProbe(t *testing.T) {
	root := t.TempDir()
	fs := afero.NewOsFs()

	t.Logf("CaseInsensitive(%s) = %v", root, CaseInsensitive(fs, root))
	assertNoProbeLeft(t, fs, root)
}

func assertNoProbeLeft(t *testing.T, fs afero.Fs, root string) {
	t.Helper()
	entries, err := afero.ReadDir(fs, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "fb-case-probe-") {
			t.Errorf("probe file left behind: %s", e.Name())
		}
	}
}

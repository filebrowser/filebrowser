package users

import (
	"path/filepath"
	"testing"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/spf13/afero"
)

// TestUserCleanFs verifies that Clean builds the user filesystem according to the
// followExternalSymlinks flag and that FullPath resolves correctly for either
// implementation.
func TestUserCleanFs(t *testing.T) {
	base := t.TempDir()
	want := filepath.Join(base, "data", "x")

	t.Run("default builds a symlink-confining ScopedFs", func(t *testing.T) {
		u := &User{Username: "u", Password: "p", Scope: "data"}
		if err := u.Clean(base, false); err != nil {
			t.Fatal(err)
		}
		if _, ok := u.Fs.(*files.ScopedFs); !ok {
			t.Fatalf("expected *files.ScopedFs, got %T", u.Fs)
		}
		if got := u.FullPath("/x"); got != want {
			t.Fatalf("FullPath: got %q, want %q", got, want)
		}
	})

	t.Run("followExternalSymlinks builds a bare BasePathFs", func(t *testing.T) {
		u := &User{Username: "u", Password: "p", Scope: "data"}
		if err := u.Clean(base, true); err != nil {
			t.Fatal(err)
		}
		if _, ok := u.Fs.(*afero.BasePathFs); !ok {
			t.Fatalf("expected *afero.BasePathFs, got %T", u.Fs)
		}
		if got := u.FullPath("/x"); got != want {
			t.Fatalf("FullPath: got %q, want %q", got, want)
		}
	})
}

package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestWithinScope(t *testing.T) {
	t.Run("non-scoped filesystem is a no-op", func(t *testing.T) {
		ok, err := WithinScope(afero.NewOsFs(), "/anything")
		if err != nil || !ok {
			t.Fatalf("expected (true, nil), got (%v, %v)", ok, err)
		}
	})

	t.Run("path inside a nested scope is allowed", func(t *testing.T) {
		scope := t.TempDir()
		if err := os.WriteFile(filepath.Join(scope, "file.txt"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		bfs := afero.NewBasePathFs(afero.NewOsFs(), scope)

		ok, err := WithinScope(bfs, "/file.txt")
		if err != nil || !ok {
			t.Fatalf("expected (true, nil), got (%v, %v)", ok, err)
		}
	})

	t.Run("new file inside scope is allowed", func(t *testing.T) {
		scope := t.TempDir()
		bfs := afero.NewBasePathFs(afero.NewOsFs(), scope)

		ok, err := WithinScope(bfs, "/does-not-exist-yet.txt")
		if err != nil || !ok {
			t.Fatalf("expected (true, nil), got (%v, %v)", ok, err)
		}
	})

	// Regression for #5975: when the scope resolves to the filesystem root,
	// root+separator used to be "//", which no path matched, so every write
	// was rejected with os.ErrPermission (HTTP 403).
	t.Run("filesystem root scope allows writes", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "file.txt")
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		bfs := afero.NewBasePathFs(afero.NewOsFs(), "/")

		ok, err := WithinScope(bfs, f)
		if err != nil || !ok {
			t.Fatalf("expected (true, nil) for a path under root scope, got (%v, %v)", ok, err)
		}
	})

	t.Run("sibling of a nested scope is rejected", func(t *testing.T) {
		base := t.TempDir()
		scope := filepath.Join(base, "srv")
		sibling := filepath.Join(base, "srvother")
		for _, d := range []string{scope, sibling} {
			if err := os.MkdirAll(d, 0o755); err != nil {
				t.Fatal(err)
			}
		}
		// A symlink lexically inside the scope pointing at a sibling directory
		// must not be followed.
		link := filepath.Join(scope, "escape")
		if err := os.Symlink(sibling, link); err != nil {
			t.Fatal(err)
		}
		bfs := afero.NewBasePathFs(afero.NewOsFs(), scope)

		ok, err := WithinScope(bfs, "/escape")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected escaping symlink to a sibling directory to be rejected")
		}
	})

	t.Run("symlink whose target stays within scope is allowed", func(t *testing.T) {
		scope := t.TempDir()
		if err := os.MkdirAll(filepath.Join(scope, "real"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(scope, "real", "f.txt"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(scope, "real"), filepath.Join(scope, "link")); err != nil {
			t.Skipf("cannot create symlink: %v", err)
		}
		bfs := afero.NewBasePathFs(afero.NewOsFs(), scope)

		ok, err := WithinScope(bfs, "/link/f.txt")
		if err != nil || !ok {
			t.Fatalf("expected (true, nil) for an in-scope symlink target, got (%v, %v)", ok, err)
		}
	})
}

// stat must reject a regular file reached through a symlinked ancestor that
// escapes the scope (GHSA-hf77-9m7w-fq8q), while still serving in-scope files.
func TestStatRejectsLinkedAncestorEscape(t *testing.T) {
	scope := t.TempDir()
	if err := os.MkdirAll(filepath.Join(scope, "shared"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(scope, "private"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scope, "private", "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scope, "shared", "ok.txt"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(scope, "private"), filepath.Join(scope, "shared", "link")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	// Filesystem scoped to the shared directory, as a public share would be.
	bfs := afero.NewBasePathFs(afero.NewOsFs(), filepath.Join(scope, "shared"))

	if _, err := stat(&FileOptions{Fs: bfs, Path: "/link/secret.txt"}); !os.IsPermission(err) {
		t.Fatalf("expected permission error for linked-ancestor escape, got %v", err)
	}
	if _, err := stat(&FileOptions{Fs: bfs, Path: "/ok.txt"}); err != nil {
		t.Fatalf("expected in-scope file to be served, got %v", err)
	}
}

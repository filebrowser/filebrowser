package fileutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

// Copying an in-scope directory that contains a symlink whose target escapes
// the user's scope must not dereference that symlink into the destination.
// Otherwise a scoped user could exfiltrate out-of-scope file content via the
// recursive copy path (GHSA-c2gv-wf5f-hjhh, an incomplete fix of
// GHSA-239w-m3h6-ch8v).
func TestCopyDoesNotDereferenceEscapingSymlink(t *testing.T) {
	base := t.TempDir()
	scope := filepath.Join(base, "scope")
	if err := os.MkdirAll(filepath.Join(scope, "srcdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	// A secret living outside the scope.
	secret := filepath.Join(base, "secret.txt")
	if err := os.WriteFile(secret, []byte("OUT-OF-SCOPE-SECRET"), 0o644); err != nil {
		t.Fatal(err)
	}

	// An escaping symlink planted inside the user's scope.
	if err := os.Symlink(secret, filepath.Join(scope, "srcdir", "link.txt")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	afs := afero.NewBasePathFs(afero.NewOsFs(), scope)

	err := Copy(afs, "/srcdir", "/dstdir", 0o644, 0o755)
	if err == nil {
		t.Fatal("expected copy of a directory containing an escaping symlink to fail")
	}

	// The escaping symlink's target content must not have landed in scope.
	if data, readErr := afero.ReadFile(afs, "/dstdir/link.txt"); readErr == nil {
		t.Fatalf("escaping symlink was dereferenced into scope: got %q", string(data))
	}
}

// A symlink whose target stays within scope is legitimate and must still be
// copied (dereferenced) so the fix does not over-block normal usage.
func TestCopyAllowsInScopeSymlink(t *testing.T) {
	scope := t.TempDir()
	if err := os.MkdirAll(filepath.Join(scope, "srcdir", "real"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scope, "srcdir", "real", "f.txt"), []byte("in-scope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(scope, "srcdir", "real", "f.txt"), filepath.Join(scope, "srcdir", "link.txt")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	afs := afero.NewBasePathFs(afero.NewOsFs(), scope)

	if err := Copy(afs, "/srcdir", "/dstdir", 0o644, 0o755); err != nil {
		t.Fatalf("expected copy of an in-scope symlink to succeed, got: %v", err)
	}

	data, err := afero.ReadFile(afs, "/dstdir/link.txt")
	if err != nil {
		t.Fatalf("expected in-scope symlink to be copied, got: %v", err)
	}
	if string(data) != "in-scope" {
		t.Fatalf("unexpected copied content: %q", string(data))
	}
}

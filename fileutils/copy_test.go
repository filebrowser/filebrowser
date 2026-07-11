package fileutils

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/spf13/afero"
)

// failingOpenFs wraps an afero.Fs and makes Open fail for one specific path,
// while every other operation (including Stat) is delegated unchanged. It
// simulates a directory that can be stat-ed but not opened/read — for example
// an unreadable sub-directory, or one whose permissions changed or that was
// removed after its parent was listed (a TOCTOU race) — encountered during a
// recursive copy.
type failingOpenFs struct {
	afero.Fs
	failOpen string
}

func (f *failingOpenFs) Open(name string) (afero.File, error) {
	if path.Clean(name) == path.Clean(f.failOpen) {
		return nil, os.ErrPermission
	}
	return f.Fs.Open(name)
}

// CopyDir is documented to keep going when it hits an error and to report the
// error afterwards. A sub-directory that cannot be opened must therefore yield
// an error (and leave the other, readable entries copied) rather than
// panicking on a nil directory handle.
func TestCopyDirUnreadableSubdirReturnsError(t *testing.T) {
	mem := afero.NewMemMapFs()
	if err := mem.MkdirAll("/srcdir/sub", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := afero.WriteFile(mem, "/srcdir/ok.txt", []byte("readable"), 0o644); err != nil {
		t.Fatal(err)
	}

	afs := &failingOpenFs{Fs: mem, failOpen: "/srcdir/sub"}

	err := Copy(afs, "/srcdir", "/dstdir", 0o644, 0o755)
	if err == nil {
		t.Fatal("expected an error when a sub-directory cannot be opened")
	}

	// The readable sibling must still have been copied (continue-on-error).
	data, readErr := afero.ReadFile(afs, "/dstdir/ok.txt")
	if readErr != nil {
		t.Fatalf("readable sibling was not copied: %v", readErr)
	}
	if string(data) != "readable" {
		t.Fatalf("unexpected copied content: %q", string(data))
	}
}

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

	afs := files.NewScopedFs(afero.NewOsFs(), scope)

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

	afs := files.NewScopedFs(afero.NewOsFs(), scope)

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

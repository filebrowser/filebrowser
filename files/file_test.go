package files

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestScopedFs(t *testing.T) {
	t.Run("path inside scope is allowed", func(t *testing.T) {
		scope := t.TempDir()
		if err := os.WriteFile(filepath.Join(scope, "file.txt"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		fs := NewScopedFs(afero.NewOsFs(), scope)

		if _, err := fs.Stat("/file.txt"); err != nil {
			t.Fatalf("expected in-scope file to be accessible, got %v", err)
		}
	})

	t.Run("new file inside scope can be created", func(t *testing.T) {
		scope := t.TempDir()
		fs := NewScopedFs(afero.NewOsFs(), scope)

		f, err := fs.OpenFile("/does-not-exist-yet.txt", os.O_RDWR|os.O_CREATE, 0o644)
		if err != nil {
			t.Fatalf("expected to create a new in-scope file, got %v", err)
		}
		_ = f.Close()
	})

	// Regression for #5975: when the scope resolves to the filesystem root,
	// root+separator used to be "//", which no path matched, so every write
	// was rejected with os.ErrPermission (HTTP 403).
	t.Run("filesystem root scope allows access", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "file.txt")
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		fs := NewScopedFs(afero.NewOsFs(), "/")

		if _, err := fs.Stat(f); err != nil {
			t.Fatalf("expected a path under root scope to be accessible, got %v", err)
		}
	})

	t.Run("escaping symlink to a sibling is rejected", func(t *testing.T) {
		base := t.TempDir()
		scope := filepath.Join(base, "srv")
		sibling := filepath.Join(base, "srvother")
		for _, d := range []string{scope, sibling} {
			if err := os.MkdirAll(d, 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(sibling, "secret.txt"), []byte("secret"), 0o644); err != nil {
			t.Fatal(err)
		}
		// A symlink lexically inside the scope pointing at a sibling directory
		// must not be followed for reads or stats.
		if err := os.Symlink(sibling, filepath.Join(scope, "escape")); err != nil {
			t.Skipf("cannot create symlink: %v", err)
		}
		fs := NewScopedFs(afero.NewOsFs(), scope)

		if _, err := fs.Stat("/escape"); !os.IsPermission(err) {
			t.Fatalf("expected stat of escaping symlink to be rejected, got %v", err)
		}
		if _, err := fs.Open("/escape/secret.txt"); !os.IsPermission(err) {
			t.Fatalf("expected read through escaping symlink to be rejected, got %v", err)
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
		fs := NewScopedFs(afero.NewOsFs(), scope)

		if _, err := fs.Stat("/link/f.txt"); err != nil {
			t.Fatalf("expected in-scope symlink target to be accessible, got %v", err)
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
	bfs := NewScopedFs(afero.NewOsFs(), filepath.Join(scope, "shared"))

	if _, err := stat(&FileOptions{Fs: bfs, Path: "/link/secret.txt"}); !os.IsPermission(err) {
		t.Fatalf("expected permission error for linked-ancestor escape, got %v", err)
	}
	if _, err := stat(&FileOptions{Fs: bfs, Path: "/ok.txt"}); err != nil {
		t.Fatalf("expected in-scope file to be served, got %v", err)
	}
}

type allowAllChecker struct{}

func (allowAllChecker) Check(string) bool {
	return true
}

type inaccessibleChildFs struct {
	afero.Fs
	child string
}

func (fs inaccessibleChildFs) Open(name string) (afero.File, error) {
	file, err := fs.Fs.Open(name)
	if err != nil {
		return nil, err
	}

	if path.Clean(name) == "/" {
		return inaccessibleChildDir{File: file}, nil
	}

	return file, nil
}

func (fs inaccessibleChildFs) Stat(name string) (os.FileInfo, error) {
	if path.Clean(name) == fs.child {
		return nil, os.ErrPermission
	}

	return fs.Fs.Stat(name)
}

func (fs inaccessibleChildFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	if path.Clean(name) == fs.child {
		return nil, false, os.ErrPermission
	}

	if lstater, ok := fs.Fs.(afero.Lstater); ok {
		return lstater.LstatIfPossible(name)
	}

	info, err := fs.Fs.Stat(name)
	return info, false, err
}

type inaccessibleChildDir struct {
	afero.File
}

func (dir inaccessibleChildDir) Readdir(int) ([]os.FileInfo, error) {
	return nil, os.ErrPermission
}

func TestReadListingSkipsInaccessibleChildren(t *testing.T) {
	memFs := afero.NewMemMapFs()
	for _, dir := range []string{"/media", "/proton-mount"} {
		if err := memFs.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	file, err := NewFileInfo(&FileOptions{
		Fs:      inaccessibleChildFs{Fs: memFs, child: "/proton-mount"},
		Path:    "/",
		Expand:  true,
		Checker: allowAllChecker{},
	})
	if err != nil {
		t.Fatal(err)
	}

	if file.Listing == nil {
		t.Fatal("expected root listing")
	}

	if got := len(file.Items); got != 1 {
		t.Fatalf("expected one accessible child, got %d", got)
	}

	if got := file.Items[0].Name; got != "media" {
		t.Fatalf("expected accessible child to be listed, got %q", got)
	}

	if got := file.NumDirs; got != 1 {
		t.Fatalf("expected one listed directory, got %d", got)
	}
}

func TestFileInfoRealPathUsesScopedFsRealPath(t *testing.T) {
	root := t.TempDir()
	file := &FileInfo{
		Fs:   NewScopedFs(afero.NewOsFs(), root),
		Path: "/root/downloads",
	}

	got := file.RealPath()
	want := filepath.Join(root, "root", "downloads")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

type openCountingFs struct {
	afero.Fs
	opens map[string]int
}

func (fs *openCountingFs) Open(name string) (afero.File, error) {
	fs.opens[path.Clean(name)]++
	return fs.Fs.Open(name)
}

func TestDetectTypeSkipsHeaderReadForKnownImageExtension(t *testing.T) {
	memFs := afero.NewMemMapFs()
	if err := afero.WriteFile(memFs, "/photo.jpg", []byte("header not needed"), 0o644); err != nil {
		t.Fatal(err)
	}

	countingFs := &openCountingFs{Fs: memFs, opens: map[string]int{}}
	file, err := NewFileInfo(&FileOptions{
		Fs:         countingFs,
		Path:       "/photo.jpg",
		Expand:     true,
		ReadHeader: true,
		Checker:    allowAllChecker{},
	})
	if err != nil {
		t.Fatal(err)
	}

	if file.Type != "image" {
		t.Fatalf("expected image type, got %q", file.Type)
	}
	if got := countingFs.opens["/photo.jpg"]; got != 0 {
		t.Fatalf("expected known image extension to avoid header reads, got %d opens", got)
	}
}

func TestDetectTypeStillReadsHeaderForUnknownExtension(t *testing.T) {
	memFs := afero.NewMemMapFs()
	if err := afero.WriteFile(memFs, "/file.unknown", []byte("plain text"), 0o644); err != nil {
		t.Fatal(err)
	}

	countingFs := &openCountingFs{Fs: memFs, opens: map[string]int{}}
	file, err := NewFileInfo(&FileOptions{
		Fs:         countingFs,
		Path:       "/file.unknown",
		Modify:     true,
		Expand:     true,
		ReadHeader: true,
		Checker:    allowAllChecker{},
	})
	if err != nil {
		t.Fatal(err)
	}

	if file.Type != "text" {
		t.Fatalf("expected text type, got %q", file.Type)
	}
	if got := countingFs.opens["/file.unknown"]; got != 1 {
		t.Fatalf("expected unknown extension to read one header, got %d opens", got)
	}
}

func TestReadListingCalculatesImageResolutionOnce(t *testing.T) {
	memFs := afero.NewMemMapFs()
	file, err := memFs.Create("/photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(0, 0, color.White)
	if err := jpeg.Encode(file, img, nil); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	countingFs := &openCountingFs{Fs: memFs, opens: map[string]int{}}
	listing, err := NewFileInfo(&FileOptions{
		Fs:         countingFs,
		Path:       "/",
		Expand:     true,
		ReadHeader: true,
		CalcImgRes: true,
		Checker:    allowAllChecker{},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(listing.Items) != 1 {
		t.Fatalf("expected one listing item, got %d", len(listing.Items))
	}
	if listing.Items[0].Resolution == nil {
		t.Fatal("expected image resolution to be populated")
	}
	if got := countingFs.opens["/photo.jpg"]; got != 1 {
		t.Fatalf("expected one image open for resolution calculation, got %d", got)
	}
}

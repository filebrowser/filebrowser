package fbhttp

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asdine/storm/v3"
	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
)

// symlinkShareStorage builds a storage whose single user is rooted at a real
// on-disk scope containing a public share "/shared" with a symlinked
// descendant "link -> ../private". Skips the test if symlinks are unavailable.
func symlinkShareStorage(t *testing.T) *storage.Storage {
	t.Helper()
	scope := t.TempDir()
	if err := os.MkdirAll(filepath.Join(scope, "shared"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(scope, "private"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scope, "private", "secret.txt"), []byte("symlink-secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(scope, "private"), filepath.Join(scope, "shared", "link")); err != nil {
		t.Skipf("cannot create symlink on this platform: %v", err)
	}

	db, err := storm.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	st, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("failed to get storage: %v", err)
	}
	if err := st.Share.Save(&share.Link{Hash: "h", UserID: 1, Path: "/shared"}); err != nil {
		t.Fatalf("failed to save share: %v", err)
	}
	if err := st.Users.Save(&users.User{
		Username: "username",
		Password: "pw",
		Perm:     users.Permissions{Share: true, Download: true},
	}); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}
	if err := st.Settings.Save(&settings.Settings{Key: []byte("key")}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}
	st.Users = &customFSUser{
		Store: st.Users,
		fs:    afero.NewBasePathFs(afero.NewOsFs(), scope),
	}
	return st
}

// Reproduces GHSA-hf77-9m7w-fq8q: a public directory share whose subtree
// contains a symlink to a directory outside the share. Requesting a regular
// file behind that linked ancestor must NOT disclose its contents.
func TestPublicShareSymlinkDescendantDisclosure(t *testing.T) {
	cases := map[string]struct {
		handler handleFunc
		path    string
	}{
		"direct file download via dl handler": {handler: publicDlHandler, path: "h/link/secret.txt"},
		"share info via share handler":        {handler: publicShareHandler, path: "h/link/secret.txt"},
		"listing of linked dir":               {handler: publicShareHandler, path: "h/link/"},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			st := symlinkShareStorage(t)

			req := newHTTPRequest(t, func(r *http.Request) { r.URL.Path = tc.path })
			recorder := httptest.NewRecorder()
			handler := handle(tc.handler, "", st, &settings.Server{})
			handler.ServeHTTP(recorder, req)

			result := recorder.Result()
			defer result.Body.Close()
			body, _ := io.ReadAll(result.Body)

			t.Logf("status=%d body=%q", result.StatusCode, string(body))
			if result.StatusCode == http.StatusOK {
				t.Errorf("VULNERABLE: leaked path outside share (status 200, body=%q)", string(body))
			}
		})
	}
}

// The listing of the public share root must omit the escaping symlink "link"
// entirely (no target metadata leak).
func TestPublicShareSymlinkListingOmitsEscapingLink(t *testing.T) {
	st := symlinkShareStorage(t)

	req := newHTTPRequest(t, func(r *http.Request) { r.URL.Path = "h/" })
	recorder := httptest.NewRecorder()
	handler := handle(publicShareHandler, "", st, &settings.Server{})
	handler.ServeHTTP(recorder, req)

	result := recorder.Result()
	defer result.Body.Close()
	body, _ := io.ReadAll(result.Body)
	if result.StatusCode != http.StatusOK {
		t.Fatalf("share root listing failed: status=%d body=%q", result.StatusCode, string(body))
	}
	if strings.Contains(string(body), "\"link\"") {
		t.Errorf("VULNERABLE: listing exposes escaping symlink: %s", string(body))
	}
}

// Reproduces the archive variant of GHSA-hf77-9m7w-fq8q: downloading the whole
// public share as a zip must not pull in files reached through a symlinked
// descendant.
func TestPublicShareSymlinkArchiveDisclosure(t *testing.T) {
	st := symlinkShareStorage(t)

	// Request the whole share root as an archive.
	req := newHTTPRequest(t, func(r *http.Request) { r.URL.Path = "h/" })
	recorder := httptest.NewRecorder()
	handler := handle(publicDlHandler, "", st, &settings.Server{})
	handler.ServeHTTP(recorder, req)

	result := recorder.Result()
	defer result.Body.Close()
	body, _ := io.ReadAll(result.Body)
	if result.StatusCode != http.StatusOK {
		t.Fatalf("archive request failed: status=%d body=%q", result.StatusCode, string(body))
	}

	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatalf("failed to read zip: %v", err)
	}
	for _, f := range zr.File {
		if strings.Contains(f.Name, "secret.txt") {
			t.Errorf("VULNERABLE: archive includes file behind symlinked descendant: %q", f.Name)
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		content, _ := io.ReadAll(rc)
		rc.Close()
		if bytes.Contains(content, []byte("symlink-secret")) {
			t.Errorf("VULNERABLE: archive entry %q leaks out-of-scope content", f.Name)
		}
	}
}

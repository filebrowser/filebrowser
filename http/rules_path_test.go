package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/filebrowser/filebrowser/v2/diskcache"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
)

// denyRuleStorage builds a scoped storage whose global settings deny denyPath.
func denyRuleStorage(t *testing.T, userScope, denyPath string, perm users.Permissions, key []byte) *storage.Storage {
	t.Helper()
	st := scopedUserStorage(t, userScope, perm, key)
	if err := st.Settings.Save(&settings.Settings{
		Key:   key,
		Rules: []rules.Rule{{Path: denyPath, Allow: false}},
	}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}
	return st
}

// Regression for GHSA-fgm5-pw99-w2p7: on a case-insensitive filesystem
// /Secret.txt and /secret.TXT name the same file, so a deny rule for one must
// cover the other. The rule check runs before any filesystem access, so the 403
// is asserted on every platform; the flag is what does the work, which the
// case-sensitive half of the test pins down.
func TestRuleDeniesCaseVariantWhenFsIsCaseInsensitive(t *testing.T) {
	userScope := t.TempDir()
	if err := os.WriteFile(filepath.Join(userScope, "Secret.txt"), []byte("SECRET"), 0o644); err != nil {
		t.Fatal(err)
	}

	key := []byte("test-signing-key")
	perm := users.Permissions{Download: true}
	st := denyRuleStorage(t, userScope, "/Secret.txt", perm, key)
	signed := signToken(t, perm, key)

	get := func(path string, caseInsensitive bool) *httptest.ResponseRecorder {
		req, _ := http.NewRequest(http.MethodGet, path, http.NoBody)
		req.Header.Set("X-Auth", signed)
		rec := httptest.NewRecorder()
		handle(rawHandler, "", st, &settings.Server{CaseInsensitiveFs: caseInsensitive}).ServeHTTP(rec, req)
		return rec
	}

	t.Run("canonical path is denied either way", func(t *testing.T) {
		for _, caseInsensitive := range []bool{false, true} {
			if rec := get("/Secret.txt", caseInsensitive); rec.Code != http.StatusForbidden {
				t.Errorf("GET /Secret.txt (caseInsensitive=%v) = %d; want 403", caseInsensitive, rec.Code)
			}
		}
	})

	t.Run("case variant is denied on a case-insensitive filesystem", func(t *testing.T) {
		if rec := get("/secret.TXT", true); rec.Code != http.StatusForbidden {
			t.Errorf("VULNERABLE: GET /secret.TXT = %d, body=%q; want 403", rec.Code, rec.Body.String())
		}
	})

	t.Run("case variant is a distinct path on a case-sensitive filesystem", func(t *testing.T) {
		if rec := get("/secret.TXT", false); rec.Code == http.StatusForbidden {
			t.Error("GET /secret.TXT = 403 with folding off; the rule must not widen on a case-sensitive filesystem")
		}
	})
}

// Regression for GHSA-fgm5-pw99-w2p7: a path is canonicalized before it is
// matched against the rules, so a traversal sequence cannot reach a denied file
// by spelling its way there. gorilla/mux cleans forward-slash traversal before
// routing, so this is defense in depth on POSIX; on Windows the equivalent
// backslash form reaches the handler intact, which cleanSeparators covers.
func TestRuleDeniesTraversalToDeniedPath(t *testing.T) {
	userScope := t.TempDir()
	if err := os.MkdirAll(filepath.Join(userScope, "allow"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userScope, "Secret.txt"), []byte("SECRET"), 0o644); err != nil {
		t.Fatal(err)
	}

	key := []byte("test-signing-key")
	perm := users.Permissions{Download: true}
	st := denyRuleStorage(t, userScope, "/Secret.txt", perm, key)

	req, _ := http.NewRequest(http.MethodGet, "/allow/../Secret.txt", http.NoBody)
	req.Header.Set("X-Auth", signToken(t, perm, key))
	rec := httptest.NewRecorder()
	handle(rawHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("VULNERABLE: GET /allow/../Secret.txt = %d, body=%q; want 403", rec.Code, rec.Body.String())
	}
}

// Canonicalizing the request path must not drop a trailing separator: POST
// distinguishes "/dir/" (create a directory) from "/dir" (write a file), and
// PUT rejects the directory form outright.
func TestCanonicalizeRequestPathKeepsTrailingSlash(t *testing.T) {
	userScope := t.TempDir()

	key := []byte("test-signing-key")
	perm := users.Permissions{Create: true, Modify: true}
	st := scopedUserStorage(t, userScope, perm, key)
	signed := signToken(t, perm, key)

	t.Run("post creates a directory", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/newdir/", http.NoBody)
		req.Header.Set("X-Auth", signed)
		rec := httptest.NewRecorder()
		handle(resourcePostHandler(diskcache.NewNoOp()), "", st, &settings.Server{}).ServeHTTP(rec, req)

		info, err := os.Stat(filepath.Join(userScope, "newdir"))
		if err != nil {
			t.Fatalf("POST /newdir/ = %d, body=%q; directory not created: %v", rec.Code, rec.Body.String(), err)
		}
		if !info.IsDir() {
			t.Error("POST /newdir/ created a file, not a directory")
		}
	})

	t.Run("put rejects a directory path", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/newdir/", http.NoBody)
		req.Header.Set("X-Auth", signed)
		rec := httptest.NewRecorder()
		handle(resourcePutHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("PUT /newdir/ = %d; want 405", rec.Code)
		}
	})
}

package fbhttp

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// Regression for the archive backslash-to-slash zip-slip (GHSA-83xp-526h-j3ww):
// a single in-scope file whose name contains backslashes is a legal POSIX
// filename, not a traversal. The archive builder must never rewrite "\" into the
// path separator "/", which would manufacture an entry like "../../evil.sh" that
// escapes the extraction directory on the downloader's machine.
func TestRawArchiveDoesNotManufactureTraversal(t *testing.T) {
	root := t.TempDir()
	userScope := filepath.Join(root, "user")
	if err := os.MkdirAll(filepath.Join(userScope, "ziptest"), 0o755); err != nil {
		t.Fatal(err)
	}

	// One legal Linux/macOS filename whose bytes include backslashes. It does not
	// traverse on the server; it only becomes "../../evil.sh" if the builder
	// turns "\" into "/".
	planted := filepath.Join(userScope, "ziptest", "..\\..\\evil.sh")
	if err := os.WriteFile(planted, []byte("#!/bin/sh\necho PWNED"), 0o644); err != nil {
		t.Skipf("cannot create backslash-named file: %v", err)
	}

	key := []byte("test-signing-key")
	perm := users.Permissions{Download: true}
	st := scopedUserStorage(t, userScope, perm, key)
	signed := signToken(t, perm, key)

	req, _ := http.NewRequest(http.MethodGet, "/ziptest?algo=zip", http.NoBody)
	req.Header.Set("X-Auth", signed)
	rec := httptest.NewRecorder()
	handle(rawHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rec.Code, rec.Body.String())
	}

	zr, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
	if err != nil {
		t.Fatalf("failed to read zip: %v", err)
	}
	if len(zr.File) == 0 {
		t.Fatal("archive has no entries")
	}

	for _, f := range zr.File {
		// The entry must be a normalized, root-relative path: no ".." segments
		// and no leading "/". Note a name may legitimately contain ".." as part of
		// a single filename (e.g. ".._.._evil.sh"), which Clean leaves untouched —
		// so compare against the normalized form rather than searching for "..".
		if strings.HasPrefix(f.Name, "/") || path.Clean("/"+f.Name) != "/"+f.Name {
			t.Errorf("VULNERABLE: archive entry escapes root: %q", f.Name)
		}
	}
}

func TestSetContentDisposition(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		filename string
		inline   bool
		expected string
	}{
		"inline simple filename": {
			filename: "document.pdf",
			inline:   true,
			expected: "inline; filename*=utf-8''" + url.PathEscape("document.pdf"),
		},
		"attachment simple filename": {
			filename: "document.pdf",
			inline:   false,
			expected: "attachment; filename*=utf-8''" + url.PathEscape("document.pdf"),
		},
		"inline non-ASCII filename": {
			filename: "日本語.txt",
			inline:   true,
			expected: "inline; filename*=utf-8''" + url.PathEscape("日本語.txt"),
		},
		"attachment non-ASCII filename": {
			filename: "日本語.txt",
			inline:   false,
			expected: "attachment; filename*=utf-8''" + url.PathEscape("日本語.txt"),
		},
		"inline filename with spaces": {
			filename: "my file.txt",
			inline:   true,
			expected: "inline; filename*=utf-8''" + url.PathEscape("my file.txt"),
		},
		"attachment filename with spaces": {
			filename: "my file.txt",
			inline:   false,
			expected: "attachment; filename*=utf-8''" + url.PathEscape("my file.txt"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/test", http.NoBody)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			if tc.inline {
				req.URL.RawQuery = "inline=true"
			}

			file := &files.FileInfo{Name: tc.filename}

			setContentDisposition(recorder, req, file)

			got := recorder.Header().Get("Content-Disposition")
			if got != tc.expected {
				t.Errorf("Content-Disposition = %q, want %q", got, tc.expected)
			}

			contentType := recorder.Header().Get("Content-Type")
			if tc.inline && contentType != "" {
				t.Errorf("Content-Type = %q, want empty", contentType)
			}
			if !tc.inline && contentType != "application/octet-stream" {
				t.Errorf("Content-Type = %q, want application/octet-stream", contentType)
			}
		})
	}
}

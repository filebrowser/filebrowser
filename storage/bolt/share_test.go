package bolt

import (
	"os"
	"sort"
	"testing"

	"github.com/asdine/storm/v3"

	"github.com/filebrowser/filebrowser/v2/share"
)

func newTestShareBackend(t *testing.T) shareBackend {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "shares-*.db")
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	_ = f.Close()

	db, err := storm.Open(f.Name())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return shareBackend{db: db}
}

func remainingHashes(t *testing.T, s shareBackend) []string {
	t.Helper()

	links, err := s.All()
	if err != nil {
		t.Fatalf("All returned error: %v", err)
	}

	hashes := make([]string, 0, len(links))
	for _, link := range links {
		hashes = append(hashes, link.Hash)
	}
	sort.Strings(hashes)
	return hashes
}

func TestDeleteWithPathPrefix(t *testing.T) {
	t.Parallel()

	s := newTestShareBackend(t)

	links := []*share.Link{
		// user 1's links
		{Hash: "u1-a", Path: "/a", UserID: 1},
		{Hash: "u1-a-child", Path: "/a/child.txt", UserID: 1},
		{Hash: "u1-abc", Path: "/abc", UserID: 1}, // not a descendant of /a
		{Hash: "u1-other", Path: "/other", UserID: 1},
		// user 2's links — must never be touched when user 1 deletes
		{Hash: "u2-a", Path: "/a", UserID: 2},
		{Hash: "u2-a-child", Path: "/a/child.txt", UserID: 2},
	}
	for _, l := range links {
		if err := s.Save(l); err != nil {
			t.Fatalf("failed to save link %s: %v", l.Hash, err)
		}
	}

	// User 1 deletes their directory /a. Only user 1's /a and its descendants
	// should be removed; /abc (sibling sharing a byte prefix) and all of user
	// 2's links must remain.
	if err := s.DeleteWithPathPrefix("/a", 1); err != nil {
		t.Fatalf("DeleteWithPathPrefix returned error: %v", err)
	}

	got := remainingHashes(t, s)
	want := []string{"u1-abc", "u1-other", "u2-a", "u2-a-child"}
	if len(got) != len(want) {
		t.Fatalf("remaining hashes = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("remaining hashes = %v, want %v", got, want)
		}
	}
}

// Regression for the trailing-slash delete leaving a stale share
// (GHSA-pp88-jhwj-5qh5): deleting "/a/" must remove the exact "/a" share and its
// descendants, not just the descendants. Siblings and other users are untouched.
func TestDeleteWithPathPrefixTrailingSlash(t *testing.T) {
	t.Parallel()

	s := newTestShareBackend(t)

	links := []*share.Link{
		{Hash: "u1-a", Path: "/a", UserID: 1},
		{Hash: "u1-a-child", Path: "/a/child.txt", UserID: 1},
		{Hash: "u1-abc", Path: "/abc", UserID: 1}, // sibling sharing a byte prefix
		{Hash: "u2-a", Path: "/a", UserID: 2},      // other user, must remain
	}
	for _, l := range links {
		if err := s.Save(l); err != nil {
			t.Fatalf("failed to save link %s: %v", l.Hash, err)
		}
	}

	// Delete with a trailing slash, as the resource delete handler does for a
	// directory request like DELETE /api/resources/a/.
	if err := s.DeleteWithPathPrefix("/a/", 1); err != nil {
		t.Fatalf("DeleteWithPathPrefix returned error: %v", err)
	}

	got := remainingHashes(t, s)
	want := []string{"u1-abc", "u2-a"}
	if len(got) != len(want) {
		t.Fatalf("remaining hashes = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("remaining hashes = %v, want %v", got, want)
		}
	}
}

func TestDeleteWithPathPrefixNoMatch(t *testing.T) {
	t.Parallel()

	s := newTestShareBackend(t)

	// No links exist at all: the storm Prefix query returns ErrNotFound, which
	// must be treated as a no-op rather than surfaced as an error.
	if err := s.DeleteWithPathPrefix("/a", 1); err != nil {
		t.Fatalf("DeleteWithPathPrefix on empty store returned error: %v", err)
	}
}

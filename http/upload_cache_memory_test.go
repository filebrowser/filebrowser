package fbhttp

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// On expiry the memory upload cache must delete an abandoned upload through the
// registered scoped removal callback, not with a raw os.Remove on the cached
// path. This test registers an entry keyed by a real file's path whose callback
// does not delete the file; if the old raw os.Remove behaviour were still in
// place the file (whose path is the key) would be removed.
func TestMemoryUploadCacheEvictionUsesRemovalCallback(t *testing.T) {
	c := newMemoryUploadCache()
	t.Cleanup(c.Close)

	f := filepath.Join(t.TempDir(), "partial.upload")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	var once sync.Once
	// Deliberately do NOT delete the file here.
	c.cache.Set(f, memoryUploadEntry{size: 1, remove: func() error {
		once.Do(func() { close(done) })
		return nil
	}}, time.Millisecond)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("eviction did not invoke the removal callback")
	}

	if _, err := os.Stat(f); err != nil {
		t.Fatalf("file at the cache key was deleted, so eviction bypassed the callback (raw os.Remove?): %v", err)
	}
}

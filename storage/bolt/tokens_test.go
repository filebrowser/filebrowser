package bolt

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/asdine/storm/v3"

	"github.com/filebrowser/filebrowser/v2/auth"
	fberrors "github.com/filebrowser/filebrowser/v2/errors"
)

func newTestTokenStore(t *testing.T) auth.TokenStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := storm.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	})
	return tokenBackend{db: db}
}

func TestTokenStore_SaveAndGet(t *testing.T) {
	t.Parallel()
	store := newTestTokenStore(t)

	token := &auth.Token{
		Token:     "test-token-123",
		UserID:    1,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := store.Save(token); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := store.Get("test-token-123")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if got.Token != token.Token {
		t.Errorf("Token = %q, want %q", got.Token, token.Token)
	}
	if got.UserID != token.UserID {
		t.Errorf("UserID = %d, want %d", got.UserID, token.UserID)
	}
}

func TestTokenStore_GetNotFound(t *testing.T) {
	t.Parallel()
	store := newTestTokenStore(t)

	_, err := store.Get("nonexistent")
	if err != fberrors.ErrNotExist {
		t.Errorf("Get() error = %v, want %v", err, fberrors.ErrNotExist)
	}
}

func TestTokenStore_Delete(t *testing.T) {
	t.Parallel()
	store := newTestTokenStore(t)

	token := &auth.Token{
		Token:     "to-delete",
		UserID:    1,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}
	if err := store.Save(token); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := store.Delete("to-delete"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err := store.Get("to-delete")
	if err != fberrors.ErrNotExist {
		t.Errorf("Get() after Delete() error = %v, want %v", err, fberrors.ErrNotExist)
	}
}

func TestTokenStore_DeleteNonexistent(t *testing.T) {
	t.Parallel()
	store := newTestTokenStore(t)

	// Deleting a nonexistent token should not error
	if err := store.Delete("nonexistent"); err != nil {
		t.Errorf("Delete() of nonexistent token error: %v", err)
	}
}

func TestTokenStore_DeleteByUser(t *testing.T) {
	t.Parallel()
	store := newTestTokenStore(t)

	// Save tokens for two different users
	for _, tok := range []*auth.Token{
		{Token: "user1-tok1", UserID: 1, ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now()},
		{Token: "user1-tok2", UserID: 1, ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now()},
		{Token: "user2-tok1", UserID: 2, ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now()},
	} {
		if err := store.Save(tok); err != nil {
			t.Fatalf("Save() error: %v", err)
		}
	}

	// Delete all tokens for user 1
	if err := store.DeleteByUser(1); err != nil {
		t.Fatalf("DeleteByUser() error: %v", err)
	}

	// User 1 tokens should be gone
	if _, err := store.Get("user1-tok1"); err != fberrors.ErrNotExist {
		t.Errorf("user1-tok1 should be deleted, got err: %v", err)
	}
	if _, err := store.Get("user1-tok2"); err != fberrors.ErrNotExist {
		t.Errorf("user1-tok2 should be deleted, got err: %v", err)
	}

	// User 2 token should still exist
	if _, err := store.Get("user2-tok1"); err != nil {
		t.Errorf("user2-tok1 should still exist, got err: %v", err)
	}
}

func TestTokenStore_DeleteByUserNoTokens(t *testing.T) {
	t.Parallel()
	store := newTestTokenStore(t)

	// Should not error when no tokens exist for user
	if err := store.DeleteByUser(999); err != nil {
		t.Errorf("DeleteByUser() with no tokens error: %v", err)
	}
}

func TestTokenStore_DeleteExpired(t *testing.T) {
	t.Parallel()
	store := newTestTokenStore(t)

	for _, tok := range []*auth.Token{
		{Token: "expired1", UserID: 1, ExpiresAt: time.Now().Add(-1 * time.Hour), CreatedAt: time.Now()},
		{Token: "expired2", UserID: 2, ExpiresAt: time.Now().Add(-1 * time.Minute), CreatedAt: time.Now()},
		{Token: "valid1", UserID: 1, ExpiresAt: time.Now().Add(1 * time.Hour), CreatedAt: time.Now()},
		{Token: "valid2", UserID: 3, ExpiresAt: time.Now().Add(2 * time.Hour), CreatedAt: time.Now()},
	} {
		if err := store.Save(tok); err != nil {
			t.Fatalf("Save() error: %v", err)
		}
	}

	if err := store.DeleteExpired(); err != nil {
		t.Fatalf("DeleteExpired() error: %v", err)
	}

	// Expired tokens should be gone
	if _, err := store.Get("expired1"); err != fberrors.ErrNotExist {
		t.Errorf("expired1 should be deleted, got err: %v", err)
	}
	if _, err := store.Get("expired2"); err != fberrors.ErrNotExist {
		t.Errorf("expired2 should be deleted, got err: %v", err)
	}

	// Valid tokens should remain
	if _, err := store.Get("valid1"); err != nil {
		t.Errorf("valid1 should still exist, got err: %v", err)
	}
	if _, err := store.Get("valid2"); err != nil {
		t.Errorf("valid2 should still exist, got err: %v", err)
	}
}

func TestTokenStore_DeleteExpiredNoExpired(t *testing.T) {
	t.Parallel()
	store := newTestTokenStore(t)

	// Should not error when no expired tokens exist
	if err := store.DeleteExpired(); err != nil {
		t.Errorf("DeleteExpired() with no expired tokens error: %v", err)
	}
}

package users

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
)

// Interface is implemented by storage
var _ Store = &Storage{}

// slowBackend is a minimal StorageBackend whose scope lookup is deliberately
// slow, so that an unsynchronized check-then-save reliably interleaves.
type slowBackend struct {
	mu    sync.Mutex
	users []*User
	delay time.Duration
}

func (b *slowBackend) GetByScope(scope string) (*User, error) {
	time.Sleep(b.delay)

	b.mu.Lock()
	defer b.mu.Unlock()
	for _, u := range b.users {
		if strings.EqualFold(u.Scope, scope) {
			return u, nil
		}
	}
	return nil, fberrors.ErrNotExist
}

func (b *slowBackend) Save(u *User) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.users = append(b.users, u)
	return nil
}

func (b *slowBackend) GetBy(interface{}) (*User, error) { return nil, fberrors.ErrNotExist }
func (b *slowBackend) Gets() ([]*User, error)           { return b.users, nil }
func (b *slowBackend) Update(*User, ...string) error    { return nil }
func (b *slowBackend) DeleteByID(uint) error            { return nil }
func (b *slowBackend) DeleteByUsername(string) error    { return nil }
func (b *slowBackend) CountAdmins() (int, error)        { return 0, nil }

// Two usernames that normalize to the same home directory must not both be
// provisioned, even when their requests are handled concurrently: checking the
// scope and saving the user has to be a single atomic step.
func TestSaveProvisionedRejectsConcurrentScopeCollision(t *testing.T) {
	back := &slowBackend{delay: 50 * time.Millisecond}
	store := NewStorage(back)

	errs := make([]error, 2)
	var wg, ready sync.WaitGroup
	ready.Add(2)
	for i, username := range []string{"teamone-x", "teamone/x"} {
		wg.Add(1)
		go func(i int, username string) {
			defer wg.Done()
			// Only start once both goroutines are actually running, so that an
			// unsynchronized implementation is guaranteed to interleave rather
			// than to pass by scheduling luck.
			ready.Done()
			ready.Wait()
			errs[i] = store.SaveProvisioned(&User{
				Username: username,
				Password: "pw",
				Scope:    "/users/teamone-x",
			}, true)
		}(i, username)
	}
	wg.Wait()

	saved := 0
	for _, err := range errs {
		switch {
		case err == nil:
			saved++
		case !errors.Is(err, fberrors.ErrExist):
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if saved != 1 {
		t.Fatalf("VULNERABLE: %d users provisioned into /users/teamone-x, want 1", saved)
	}
	if len(back.users) != 1 {
		t.Fatalf("%d users stored, want 1", len(back.users))
	}
}

// A scope that was not derived from the username (explicitly configured by an
// administrator or returned by an auth hook) may legitimately be shared, so it
// must not be subject to the collision check.
func TestSaveProvisionedAllowsSharedExplicitScope(t *testing.T) {
	back := &slowBackend{}
	store := NewStorage(back)

	for _, username := range []string{"alice", "bob"} {
		if err := store.SaveProvisioned(&User{
			Username: username,
			Password: "pw",
			Scope:    "/shared/team",
		}, false); err != nil {
			t.Fatalf("saving %q with an explicit shared scope: %v", username, err)
		}
	}

	if len(back.users) != 2 {
		t.Fatalf("%d users stored, want 2", len(back.users))
	}
}

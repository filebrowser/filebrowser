package share

import (
	"errors"
	"testing"
	"time"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
)

func TestValidateHash(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input   string
		want    string
		wantErr error
	}{
		"valid key": {
			input: "team-docs_2026",
			want:  "team-docs_2026",
		},
		"trimmed key": {
			input: "  TeamDocs  ",
			want:  "TeamDocs",
		},
		"empty key": {
			input:   "   ",
			wantErr: fberrors.ErrInvalidRequestParams,
		},
		"space is invalid": {
			input:   "team docs",
			wantErr: fberrors.ErrInvalidRequestParams,
		},
		"slash is invalid": {
			input:   "team/docs",
			wantErr: fberrors.ErrInvalidRequestParams,
		},
		"dot is invalid": {
			input:   "team.docs",
			wantErr: fberrors.ErrInvalidRequestParams,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := ValidateHash(tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ValidateHash(%q) error = %v, want %v", tc.input, err, tc.wantErr)
			}

			if got != tc.want {
				t.Fatalf("ValidateHash(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestStorageSaveRejectsActiveDuplicateHash(t *testing.T) {
	t.Parallel()

	back := &fakeShareBackend{
		links: map[string]*Link{
			"team-docs": {
				Hash:   "team-docs",
				Path:   "/docs.txt",
				UserID: 1,
			},
		},
	}

	storage := NewStorage(back)
	err := storage.Save(&Link{
		Hash:   "team-docs",
		Path:   "/other.txt",
		UserID: 2,
	})

	if !errors.Is(err, fberrors.ErrExist) {
		t.Fatalf("Save duplicate error = %v, want %v", err, fberrors.ErrExist)
	}

	if back.saveCalls != 0 {
		t.Fatalf("expected no save call on duplicate, got %d", back.saveCalls)
	}
}

func TestStorageSaveDeletesExpiredDuplicateBeforeReuse(t *testing.T) {
	t.Parallel()

	back := &fakeShareBackend{
		links: map[string]*Link{
			"team-docs": {
				Hash:   "team-docs",
				Path:   "/old.txt",
				UserID: 1,
				Expire: time.Now().Add(-time.Hour).Unix(),
			},
		},
	}

	storage := NewStorage(back)
	err := storage.Save(&Link{
		Hash:   "team-docs",
		Path:   "/new.txt",
		UserID: 2,
	})
	if err != nil {
		t.Fatalf("Save expired duplicate returned error: %v", err)
	}

	if back.deleteCalls != 1 {
		t.Fatalf("expected 1 delete call, got %d", back.deleteCalls)
	}

	if back.saveCalls != 1 {
		t.Fatalf("expected 1 save call, got %d", back.saveCalls)
	}

	if got := back.links["team-docs"]; got == nil || got.Path != "/new.txt" || got.UserID != 2 {
		t.Fatalf("saved link = %#v, want new replacement", got)
	}
}

type fakeShareBackend struct {
	links       map[string]*Link
	saveCalls   int
	deleteCalls int
}

func (f *fakeShareBackend) All() ([]*Link, error) {
	return nil, nil
}

func (f *fakeShareBackend) FindByUserID(id uint) ([]*Link, error) {
	return nil, nil
}

func (f *fakeShareBackend) GetByHash(hash string) (*Link, error) {
	if link, ok := f.links[hash]; ok {
		copy := *link
		return &copy, nil
	}

	return nil, fberrors.ErrNotExist
}

func (f *fakeShareBackend) GetPermanent(path string, id uint) (*Link, error) {
	return nil, fberrors.ErrNotExist
}

func (f *fakeShareBackend) Gets(path string, id uint) ([]*Link, error) {
	return nil, nil
}

func (f *fakeShareBackend) Save(s *Link) error {
	f.saveCalls++
	copy := *s
	f.links[s.Hash] = &copy
	return nil
}

func (f *fakeShareBackend) Delete(hash string) error {
	f.deleteCalls++
	delete(f.links, hash)
	return nil
}

func (f *fakeShareBackend) DeleteWithPathPrefix(path string) error {
	return nil
}

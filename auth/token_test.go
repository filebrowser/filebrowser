package auth

import (
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	t.Parallel()

	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() returned error: %v", err)
	}

	if len(token) != 64 {
		t.Errorf("expected token length 64, got %d", len(token))
	}

	token2, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() returned error: %v", err)
	}
	if token == token2 {
		t.Error("two generated tokens should not be identical")
	}
}

func TestGenerateTokenUniqueness(t *testing.T) {
	t.Parallel()

	seen := make(map[string]bool)
	for range 100 {
		token, err := GenerateToken()
		if err != nil {
			t.Fatalf("GenerateToken() returned error: %v", err)
		}
		if seen[token] {
			t.Fatalf("duplicate token generated: %s", token)
		}
		seen[token] = true
	}
}

func TestTokenIsExpired(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		expiresAt time.Time
		want      bool
	}{
		"expired token": {
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		"valid token": {
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		"just expired": {
			expiresAt: time.Now().Add(-1 * time.Millisecond),
			want:      true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			token := &Token{
				Token:     "test",
				UserID:    1,
				ExpiresAt: tc.expiresAt,
				CreatedAt: time.Now(),
			}
			if got := token.IsExpired(); got != tc.want {
				t.Errorf("IsExpired() = %v, want %v", got, tc.want)
			}
		})
	}
}

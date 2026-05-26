package users

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// ValidateScopes checks that scopes don't overlap (no scope is a child of another)
// and that no two scopes share the same basename.
func ValidateScopes(scopes []string) error {
	if len(scopes) <= 1 {
		return nil
	}

	// Normalize all scopes
	normalized := make([]string, len(scopes))
	for i, s := range scopes {
		normalized[i] = path.Clean("/" + s)
	}

	// Check for duplicate basenames
	basenames := make(map[string]string, len(normalized))
	for _, s := range normalized {
		base := filepath.Base(s)
		if existing, ok := basenames[base]; ok {
			return fmt.Errorf("scopes %q and %q have the same folder name %q", existing, s, base)
		}
		basenames[base] = s
	}

	// Check for nested scopes
	for i, a := range normalized {
		for j, b := range normalized {
			if i == j {
				continue
			}
			// Check if a is a parent of b (or equal)
			aWithSlash := a
			if !strings.HasSuffix(aWithSlash, "/") {
				aWithSlash += "/"
			}
			if strings.HasPrefix(b, aWithSlash) || a == b {
				return fmt.Errorf("scope %q is nested inside %q; overlapping scopes are not allowed", b, a)
			}
		}
	}

	return nil
}

// ScopeBaseName returns the display name for a scope (its last path segment).
func ScopeBaseName(scope string) string {
	return filepath.Base(path.Clean("/" + scope))
}

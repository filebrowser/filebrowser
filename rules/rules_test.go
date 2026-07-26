package rules

import "testing"

func TestRuleMatches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		rulePath string
		testPath string
		fold     bool
		want     bool
	}{
		{"exact match", "/uploads", "/uploads", false, true},
		{"child path", "/uploads", "/uploads/file.txt", false, true},
		{"sibling prefix", "/uploads", "/uploads_backup/secret.txt", false, false},
		{"root rule", "/", "/anything", false, true},
		{"trailing slash rule", "/uploads/", "/uploads/file.txt", false, true},
		{"trailing slash no sibling", "/uploads/", "/uploads_backup/file.txt", false, false},
		{"nested child", "/data/shared", "/data/shared/docs/file.txt", false, true},
		{"nested sibling", "/data/shared", "/data/shared_private/file.txt", false, false},

		// On a case-insensitive filesystem /Secret.txt and /secret.TXT name the
		// same file, so a rule for one must cover the other. See
		// GHSA-fgm5-pw99-w2p7.
		{"case variant not folded", "/Secret.txt", "/secret.TXT", false, false},
		{"case variant folded", "/Secret.txt", "/secret.TXT", true, true},
		{"case variant child folded", "/Private", "/private/notes.txt", true, true},
		{"exact match folded", "/uploads", "/uploads", true, true},
		// Folding must not widen a rule past its own path segment.
		{"sibling prefix folded", "/uploads", "/UPLOADS_backup/secret.txt", true, false},
		{"nested sibling folded", "/data/shared", "/data/SHARED_private/x", true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := &Rule{Path: tc.rulePath}
			got := r.Matches(tc.testPath, tc.fold)
			if got != tc.want {
				t.Errorf("Rule{Path: %q}.Matches(%q, fold=%v) = %v; want %v",
					tc.rulePath, tc.testPath, tc.fold, got, tc.want)
			}
		})
	}
}

// A regex rule is authored by an administrator and means exactly what it says,
// so folding must not reach it: the admin opts in with (?i) instead.
func TestRuleMatchesRegexIgnoresFold(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw  string
		path string
		want bool
	}{
		{`^/Secret\.txt$`, "/Secret.txt", true},
		{`^/Secret\.txt$`, "/secret.TXT", false},
		{`(?i)^/Secret\.txt$`, "/secret.TXT", true},
	}

	for _, tc := range cases {
		for _, fold := range []bool{false, true} {
			r := &Rule{Regex: true, Regexp: &Regexp{Raw: tc.raw}}
			if got := r.Matches(tc.path, fold); got != tc.want {
				t.Errorf("Rule{Regexp: %q}.Matches(%q, fold=%v) = %v; want %v",
					tc.raw, tc.path, fold, got, tc.want)
			}
		}
	}
}

func TestMatchHidden(t *testing.T) {
	cases := map[string]bool{
		"/":                   false,
		"/src":                false,
		"/src/":               false,
		"/.circleci":          true,
		"/a/b/c/.docker.json": true,
		".docker.json":        true,
		"Dockerfile":          false,
		"/Dockerfile":         false,
	}

	for path, want := range cases {
		got := MatchHidden(path)
		if got != want {
			t.Errorf("MatchHidden(%s)=%v; want %v", path, got, want)
		}
	}
}

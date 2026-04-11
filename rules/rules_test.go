package rules

import "testing"

func TestRuleMatches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		rulePath string
		testPath string
		want     bool
	}{
		{"exact match", "/uploads", "/uploads", true},
		{"child path", "/uploads", "/uploads/file.txt", true},
		{"sibling prefix", "/uploads", "/uploads_backup/secret.txt", false},
		{"root rule", "/", "/anything", true},
		{"trailing slash rule", "/uploads/", "/uploads/file.txt", true},
		{"trailing slash no sibling", "/uploads/", "/uploads_backup/file.txt", false},
		{"nested child", "/data/shared", "/data/shared/docs/file.txt", true},
		{"nested sibling", "/data/shared", "/data/shared_private/file.txt", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := &Rule{Path: tc.rulePath}
			got := r.Matches(tc.testPath)
			if got != tc.want {
				t.Errorf("Rule{Path: %q}.Matches(%q) = %v; want %v", tc.rulePath, tc.testPath, got, tc.want)
			}
		})
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

package fbhttp

import "testing"

// cleanSeparators takes the host separator explicitly so the Windows behaviour
// can be asserted from any platform. See GHSA-fgm5-pw99-w2p7: on Windows a
// backslash is a path separator the filesystem resolves but the rule checker
// used to treat as an ordinary character, so "/allow\..\Secret.txt" evaded a
// rule for "/Secret.txt" and still opened it.
func TestCleanSeparators(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		sep  string
		want string
	}{
		// Windows: a backslash is a separator and must be resolved before the
		// path is matched against a rule.
		{"windows traversal", `/allow\..\Secret.txt`, `\`, "/Secret.txt"},
		{"windows leading backslash", `\Secret.txt`, `\`, "/Secret.txt"},
		{"windows mixed separators", `/a/b\..\..\c`, `\`, "/c"},
		{"windows already canonical", "/Secret.txt", `\`, "/Secret.txt"},
		{"windows relative", `allow\..\Secret.txt`, `\`, "/Secret.txt"},
		{"windows empty", "", `\`, "/"},

		// POSIX: a backslash is a legal filename character and must survive, or
		// files named with one become unreachable.
		{"posix backslash is a filename character", `/a\b.txt`, "/", `/a\b.txt`},
		{"posix traversal", "/allow/../Secret.txt", "/", "/Secret.txt"},
		{"posix relative", "Secret.txt", "/", "/Secret.txt"},
		{"posix multiple slashes", "//a///b", "/", "/a/b"},
		{"posix empty", "", "/", "/"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := cleanSeparators(tc.in, tc.sep); got != tc.want {
				t.Errorf("cleanSeparators(%q, %q) = %q; want %q", tc.in, tc.sep, got, tc.want)
			}
		})
	}
}

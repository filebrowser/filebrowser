package types

import (
	"regexp"
)

// Rule is a allow/disallow rule.
type Rule struct {
	Regex  bool
	Allow  bool
	Path   string
	Regexp *Regexp
}

// Regexp is a wrapper to the native regexp type where we
// save the raw expression.
type Regexp struct {
	Raw    string
	regexp *regexp.Regexp
}

// MatchString checks if a string matches the regexp.
func (r *Regexp) MatchString(s string) bool {
	if r.regexp == nil {
		r.regexp = regexp.MustCompile(r.Raw)
	}

	return r.regexp.MatchString(s)
}

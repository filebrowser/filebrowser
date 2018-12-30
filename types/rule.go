package types

import (
	"regexp"
	"strings"
)

// Rule is a allow/disallow rule.
type Rule struct {
	Regex  bool    `json:"regex"`
	Allow  bool    `json:"allow"`
	Path   string  `json:"path"`
	Regexp *Regexp `json:"regexp"`
}

// Regexp is a wrapper to the native regexp type where we
// save the raw expression.
type Regexp struct {
	Raw    string `json:"raw"`
	regexp *regexp.Regexp
}

// MatchString checks if a string matches the regexp.
func (r *Regexp) MatchString(s string) bool {
	if r.regexp == nil {
		r.regexp = regexp.MustCompile(r.Raw)
	}

	return r.regexp.MatchString(s)
}

func isAllowed(path string, rules []Rule) bool {
	for _, rule := range rules {
		if rule.Regex {
			if rule.Regexp.MatchString(path) {
				return rule.Allow
			}
		} else if strings.HasPrefix(path, rule.Path) {
			return rule.Allow
		}
	}

	return true
}

package search

import (
	"os"
	"strings"

	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/spf13/afero"
)

type searchOptions struct {
	CaseSensitive bool
	Conditions    []condition
	Terms         []string
}

// Search searches for a query in a fs.
func Search(fs afero.Fs, scope, query string, checker rules.Checker, found func(path string, f os.FileInfo) error) error {
	search := parseSearch(query)

	return afero.Walk(fs, scope, func(path string, f os.FileInfo, err error) error {
		path = strings.TrimPrefix(path, "/")
		path = strings.Replace(path, "\\", "/", -1)
		
		if !checker.Check(path) {
			return nil
		}

		if !search.CaseSensitive {
			path = strings.ToLower(path)
		}

		if !search.CaseSensitive {
			path = strings.ToLower(path)
		}

		if len(search.Conditions) > 0 {
			match := false

			for _, t := range search.Conditions {
				if t(path) {
					match = true
					break
				}
			}

			if !match {
				return nil
			}
		}

		if len(search.Terms) > 0 {
			for _, term := range search.Terms {
				if strings.Contains(path, term) {
					return found(strings.TrimPrefix(path, scope), f)
				}
			}
		}

		return nil
	})
}

package search

import (
	"os"
	"strings"

	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/rules"
)

type searchOptions struct {
	CaseSensitive bool
	Conditions    []condition
	Terms         []string
}

// Search searches for a query in a fs.
func Search(fs afero.Fs, scope, query string, checker rules.Checker, found func(path string, f os.FileInfo) error) error {
	search := parseSearch(query)

	scope = strings.Replace(scope, "\\", "/", -1)
	scope = strings.TrimPrefix(scope, "/")
	scope = strings.TrimSuffix(scope, "/")
	scope = "/" + scope + "/"

	return afero.Walk(fs, scope, func(originalPath string, f os.FileInfo, err error) error {
		originalPath = strings.Replace(originalPath, "\\", "/", -1)
		originalPath = strings.TrimPrefix(originalPath, "/")
		originalPath = "/" + originalPath
		path := originalPath

		if path == scope {
			return nil
		}

		if !checker.Check(path) {
			return nil
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
					return found(strings.TrimPrefix(originalPath, scope), f)
				}
			}
		}

		return nil
	})
}

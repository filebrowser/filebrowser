package search

import (
	"mime"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	typeRegexp = regexp.MustCompile(`type:(\w+)`)
)

type condition func(path string) bool

func extensionCondition(extension string) condition {
	return func(path string) bool {
		return filepath.Ext(path) == "."+extension
	}
}

func imageCondition(path string) bool {
	extension := filepath.Ext(path)
	mimetype := mime.TypeByExtension(extension)

	return strings.HasPrefix(mimetype, "image")
}

func audioCondition(path string) bool {
	extension := filepath.Ext(path)
	mimetype := mime.TypeByExtension(extension)

	return strings.HasPrefix(mimetype, "audio")
}

func videoCondition(path string) bool {
	extension := filepath.Ext(path)
	mimetype := mime.TypeByExtension(extension)

	return strings.HasPrefix(mimetype, "video")
}

func parseSearch(value string) *searchOptions {
	opts := &searchOptions{
		CaseSensitive: strings.Contains(value, "case:sensitive"),
		Conditions:    []condition{},
		Terms:         []string{},
	}

	// removes the options from the value
	value = strings.Replace(value, "case:insensitive", "", -1)
	value = strings.Replace(value, "case:sensitive", "", -1)
	value = strings.TrimSpace(value)

	types := typeRegexp.FindAllStringSubmatch(value, -1)
	for _, t := range types {
		if len(t) == 1 {
			continue
		}

		switch t[1] {
		case "image":
			opts.Conditions = append(opts.Conditions, imageCondition)
		case "audio", "music":
			opts.Conditions = append(opts.Conditions, audioCondition)
		case "video":
			opts.Conditions = append(opts.Conditions, videoCondition)
		default:
			opts.Conditions = append(opts.Conditions, extensionCondition(t[1]))
		}
	}

	if len(types) > 0 {
		// Remove the fields from the search value.
		value = typeRegexp.ReplaceAllString(value, "")
	}

	// If it's canse insensitive, put everything in lowercase.
	if !opts.CaseSensitive {
		value = strings.ToLower(value)
	}

	// Remove the spaces from the search value.
	value = strings.TrimSpace(value)

	if value == "" {
		return opts
	}

	// if the value starts with " and finishes what that character, we will
	// only search for that term
	if value[0] == '"' && value[len(value)-1] == '"' {
		unique := strings.TrimPrefix(value, "\"")
		unique = strings.TrimSuffix(unique, "\"")

		opts.Terms = []string{unique}
		return opts
	}

	opts.Terms = strings.Split(value, " ")
	return opts
}

package variables

import (
	"strings"
	"unicode"
)

var splitCapitalizeExceptions = map[string]string{
	"youtube":    "YouTube",
	"github":     "GitHub",
	"googleplus": "Google Plus",
	"linkedin":   "LinkedIn",
}

// SplitCapitalize splits a string by its uppercase letters and capitalize the
// first letter of the string
func SplitCapitalize(name string) string {
	if val, ok := splitCapitalizeExceptions[strings.ToLower(name)]; ok {
		return val
	}

	var words []string
	l := 0
	for s := name; s != ""; s = s[l:] {
		l = strings.IndexFunc(s[1:], unicode.IsUpper) + 1
		if l <= 0 {
			l = len(s)
		}
		words = append(words, s[:l])
	}

	name = ""

	for _, element := range words {
		name += element + " "
	}

	name = strings.ToLower(name[:len(name)-1])
	name = strings.ToUpper(string(name[0])) + name[1:]

	return name
}

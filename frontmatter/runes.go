package frontmatter

import (
	"bytes"
	"strings"
)

// HasRune checks if the file has the frontmatter rune
func HasRune(file []byte) bool {
	return strings.HasPrefix(string(file), "---") ||
		strings.HasPrefix(string(file), "+++") ||
		strings.HasPrefix(string(file), "{")
}

// AppendRune appends the frontmatter rune to a file
func AppendRune(frontmatter []byte, mark rune) []byte {
	frontmatter = bytes.TrimSpace(frontmatter)

	switch mark {
	case '-':
		return []byte("---\n" + string(frontmatter) + "\n---")
	case '+':
		return []byte("+++\n" + string(frontmatter) + "\n+++")
	case '{':
		return []byte("{\n" + string(frontmatter) + "\n}")
	}

	return frontmatter
}

func RuneToStringFormat(mark rune) string {
	switch mark {
	case '-':
		return "yaml"
	case '+':
		return "toml"
	case '{':
		return "json"
	default:
		return ""
	}
}

func StringFormatToRune(format string) rune {
	switch format {
	case "yaml":
		return '-'
	case "toml":
		return '+'
	case "json":
		return '{'
	default:
		return '0'
	}
}

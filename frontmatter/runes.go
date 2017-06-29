package frontmatter

import (
	"errors"
	"strings"
)

// HasRune checks if the file has the frontmatter rune
func HasRune(file string) bool {
	return strings.HasPrefix(file, "---") ||
		strings.HasPrefix(file, "+++") ||
		strings.HasPrefix(file, "{")
}

// AppendRune appends the frontmatter rune to a file
func AppendRune(frontmatter string, mark rune) string {
	frontmatter = strings.TrimSpace(frontmatter)

	switch mark {
	case '-':
		return "---\n" + frontmatter + "\n---"
	case '+':
		return "+++\n" + frontmatter + "\n+++"
	case '{':
		return "{\n" + frontmatter + "\n}"
	}

	return frontmatter
}

// RuneToStringFormat converts the rune to a string with the format
func RuneToStringFormat(mark rune) (string, error) {
	switch mark {
	case '-':
		return "yaml", nil
	case '+':
		return "toml", nil
	case '{', '}':
		return "json", nil
	default:
		return "", errors.New("Unsupported format type")
	}
}

// StringFormatToRune converts the format name to its rune
func StringFormatToRune(format string) (rune, error) {
	switch format {
	case "yaml":
		return '-', nil
	case "toml":
		return '+', nil
	case "json":
		return '{', nil
	default:
		return '0', errors.New("Unsupported format type")
	}
}

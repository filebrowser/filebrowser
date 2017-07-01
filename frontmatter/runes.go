package frontmatter

import (
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

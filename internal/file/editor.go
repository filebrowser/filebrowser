package file

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/spf13/hugo/parser"
)

// Editor contains the information for the editor page
type Editor struct {
	Class       string
	Mode        string
	Content     string
	FrontMatter *Content
}

// GetEditor gets the editor based on a FileInfo struct
func (i *Info) GetEditor() (*Editor, error) {
	// Create a new editor variable and set the mode
	editor := new(Editor)
	editor.Mode = strings.TrimPrefix(filepath.Ext(i.Name), ".")

	switch editor.Mode {
	case "md", "markdown", "mdown", "mmark":
		editor.Mode = "markdown"
	case "asciidoc", "adoc", "ad":
		editor.Mode = "asciidoc"
	case "rst":
		editor.Mode = "rst"
	case "html", "htm":
		editor.Mode = "html"
	case "js":
		editor.Mode = "javascript"
	}

	var page parser.Page
	var err error

	// Handle the content depending on the file extension
	switch editor.Mode {
	case "markdown", "asciidoc", "rst":
		if editor.hasFrontMatterRune(i.Raw) {
			// Starts a new buffer and parses the file using Hugo's functions
			buffer := bytes.NewBuffer(i.Raw)
			page, err = parser.ReadFrom(buffer)
			if err != nil {
				return editor, err
			}

			// Parses the page content and the frontmatter
			editor.Content = strings.TrimSpace(string(page.Content()))
			editor.FrontMatter, _, err = Pretty(page.FrontMatter())
			editor.Class = "complete"
		} else {
			// The editor will handle only content
			editor.Class = "content-only"
			editor.Content = i.Content
		}
	case "json", "toml", "yaml":
		// Defines the class and declares an error
		editor.Class = "frontmatter-only"

		// Checks if the file already has the frontmatter rune and parses it
		if editor.hasFrontMatterRune(i.Raw) {
			editor.FrontMatter, _, err = Pretty(i.Raw)
		} else {
			editor.FrontMatter, _, err = Pretty(editor.appendFrontMatterRune(i.Raw, editor.Mode))
		}

		// Check if there were any errors
		if err != nil {
			return editor, err
		}
	default:
		// The editor will handle only content
		editor.Class = "content-only"
		editor.Content = i.Content
	}

	return editor, nil
}

func (e Editor) hasFrontMatterRune(file []byte) bool {
	return strings.HasPrefix(string(file), "---") ||
		strings.HasPrefix(string(file), "+++") ||
		strings.HasPrefix(string(file), "{")
}

func (e Editor) appendFrontMatterRune(frontmatter []byte, language string) []byte {
	switch language {
	case "yaml":
		return []byte("---\n" + string(frontmatter) + "\n---")
	case "toml":
		return []byte("+++\n" + string(frontmatter) + "\n+++")
	case "json":
		return frontmatter
	}

	return frontmatter
}

// CanBeEdited checks if the extension of a file is supported by the editor
func CanBeEdited(filename string) bool {
	extensions := [...]string{
		"md", "markdown", "mdown", "mmark",
		"asciidoc", "adoc", "ad",
		"rst",
		".json", ".toml", ".yaml",
		".css", ".sass", ".scss",
		".js",
		".html",
		".txt",
	}

	for _, extension := range extensions {
		if strings.HasSuffix(filename, extension) {
			return true
		}
	}

	return false
}

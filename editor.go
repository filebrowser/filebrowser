package filemanager

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/frontmatter"
	"github.com/spf13/hugo/parser"
)

// Editor contains the information for the editor page
type Editor struct {
	Class       string
	Mode        string
	Content     string
	FrontMatter *frontmatter.Content
}

// GetEditor gets the editor based on a FileInfo struct
func (i *FileInfo) GetEditor() (*Editor, error) {
	// Create a new editor variable and set the mode
	editor := new(Editor)
	editor.Mode = strings.TrimPrefix(filepath.Ext(i.Name()), ".")

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
		if !hasFrontMatterRune(i.Content) {
			editor.Class = "content-only"
			editor.Content = i.StringifyContent()
			break
		}

		// Starts a new buffer and parses the file using Hugo's functions
		buffer := bytes.NewBuffer(i.Content)
		page, err = parser.ReadFrom(buffer)
		editor.Class = "complete"

		if err != nil {
			editor.Class = "content-only"
			editor.Content = i.StringifyContent()
			break
		}

		// Parses the page content and the frontmatter
		editor.Content = strings.TrimSpace(string(page.Content()))
		editor.FrontMatter, _, err = frontmatter.Pretty(page.FrontMatter())
	case "json", "toml", "yaml":
		// Defines the class and declares an error
		editor.Class = "frontmatter-only"

		// Checks if the file already has the frontmatter rune and parses it
		if hasFrontMatterRune(i.Content) {
			editor.FrontMatter, _, err = frontmatter.Pretty(i.Content)
		} else {
			editor.FrontMatter, _, err = frontmatter.Pretty(appendFrontMatterRune(i.Content, editor.Mode))
		}

		// Check if there were any errors
		if err != nil {
			editor.Class = "content-only"
			editor.Content = i.StringifyContent()
			break
		}
	default:
		editor.Class = "content-only"
		editor.Content = i.StringifyContent()
	}

	return editor, nil
}

// hasFrontMatterRune checks if the file has the frontmatter rune
func hasFrontMatterRune(file []byte) bool {
	return strings.HasPrefix(string(file), "---") ||
		strings.HasPrefix(string(file), "+++") ||
		strings.HasPrefix(string(file), "{")
}

// appendFrontMatterRune appends the frontmatter rune to a file
func appendFrontMatterRune(frontmatter []byte, language string) []byte {
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

// canBeEdited checks if the extension of a file is supported by the editor
func canBeEdited(filename string) bool {
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

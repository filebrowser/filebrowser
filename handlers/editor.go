package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/file"
	"github.com/hacdias/caddy-filemanager/frontmatter"
	"github.com/spf13/hugo/parser"
)

// Editor contains the information for the editor page
type Editor struct {
	Class       string
	Mode        string
	Visual      bool
	Content     string
	FrontMatter *frontmatter.Content
}

// GetEditor gets the editor based on a FileInfo struct
func GetEditor(r *http.Request, i *file.Info) (*Editor, error) {
	var err error

	// Create a new editor variable and set the mode
	editor := new(Editor)
	editor.Mode = editorMode(i.Name)
	editor.Class = editorClass(editor.Mode)

	if editor.Class == "frontmatter-only" || editor.Class == "complete" {
		editor.Visual = true
	}

	if r.URL.Query().Get("visual") == "false" {
		editor.Class = "content-only"
	}

	if editor.Class == "frontmatter-only" {
		// Checks if the file already has the frontmatter rune and parses it
		if frontmatter.HasRune(i.Content) {
			editor.FrontMatter, _, err = frontmatter.Pretty(i.Content)
		} else {
			editor.FrontMatter, _, err = frontmatter.Pretty(frontmatter.AppendRune(i.Content, editor.Mode))
		}
	}

	if editor.Class == "complete" && frontmatter.HasRune(i.Content) {
		var page parser.Page
		// Starts a new buffer and parses the file using Hugo's functions
		buffer := bytes.NewBuffer(i.Content)
		page, err = parser.ReadFrom(buffer)
		editor.Class = "complete"

		if err == nil {
			// Parses the page content and the frontmatter
			editor.Content = strings.TrimSpace(string(page.Content()))
			editor.FrontMatter, _, err = frontmatter.Pretty(page.FrontMatter())
		}
	}

	if editor.Class == "complete" && !frontmatter.HasRune(i.Content) {
		err = errors.New("Complete but without rune")
	}

	if editor.Class == "content-only" || err != nil {
		editor.Class = "content-only"
		editor.Content = i.StringifyContent()
	}

	return editor, nil
}

func editorClass(mode string) string {
	switch mode {
	case "json", "toml", "yaml":
		return "frontmatter-only"
	case "markdown", "asciidoc", "rst":
		return "complete"
	}

	return "content-only"
}

func editorMode(filename string) string {
	mode := strings.TrimPrefix(filepath.Ext(filename), ".")

	switch mode {
	case "md", "markdown", "mdown", "mmark":
		mode = "markdown"
	case "asciidoc", "adoc", "ad":
		mode = "asciidoc"
	case "rst":
		mode = "rst"
	case "html", "htm":
		mode = "html"
	case "js":
		mode = "javascript"
	case "go":
		mode = "golang"
	}

	return mode
}

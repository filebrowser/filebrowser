package filemanager

import (
	"bytes"
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/hacdias/filemanager/frontmatter"
	"github.com/spf13/hugo/parser"
)

// editor contains the information to fill the editor template.
type editor struct {
	Class       string
	Mode        string
	Visual      bool
	Content     string
	FrontMatter struct {
		Content *frontmatter.Content
		Rune    rune
	}
}

// getEditor gets the editor based on a Info struct
func getEditor(r *http.Request, i *fileInfo) (*editor, error) {
	var err error

	// Create a new editor variable and set the mode
	e := &editor{}
	e.Mode = editorMode(i.Name)
	e.Class = editorClass(e.Mode)

	if e.Class == "frontmatter-only" || e.Class == "complete" {
		e.Visual = true
	}

	if r.URL.Query().Get("visual") == "false" {
		e.Class = "content-only"
	}

	hasRune := frontmatter.HasRune(i.content)

	if e.Class == "frontmatter-only" && !hasRune {
		e.FrontMatter.Rune, err = frontmatter.StringFormatToRune(e.Mode)
		if err != nil {
			goto Error
		}
		i.content = frontmatter.AppendRune(i.content, e.FrontMatter.Rune)
		hasRune = true
	}

	if e.Class == "frontmatter-only" && hasRune {
		e.FrontMatter.Content, _, err = frontmatter.Pretty(i.content)
		if err != nil {
			goto Error
		}
	}

	if e.Class == "complete" && hasRune {
		var page parser.Page
		// Starts a new buffer and parses the file using Hugo's functions
		buffer := bytes.NewBuffer(i.content)
		page, err = parser.ReadFrom(buffer)

		if err != nil {
			goto Error
		}

		// Parses the page content and the frontmatter
		e.Content = strings.TrimSpace(string(page.Content()))
		e.FrontMatter.Rune = rune(i.content[0])
		e.FrontMatter.Content, _, err = frontmatter.Pretty(page.FrontMatter())
	}

	if e.Class == "complete" && !hasRune {
		err = errors.New("Complete but without rune")
	}

Error:
	if e.Class == "content-only" || err != nil {
		e.Class = "content-only"
		e.Content = i.StringifyContent()
	}

	return e, nil
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

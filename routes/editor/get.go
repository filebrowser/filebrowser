package editor

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/tools/frontmatter"
	"github.com/hacdias/caddy-hugo/tools/templates"
	"github.com/spf13/hugo/parser"
)

type editor struct {
	Name        string
	Class       string
	IsPost      bool
	Mode        string
	Content     string
	FrontMatter interface{}
	Config      *config.Config
}

// GET handles the GET method on editor page
func GET(w http.ResponseWriter, r *http.Request, c *config.Config, filename string) (int, error) {
	// Check if the file format is supported. If not, send a "Not Acceptable"
	// header and an error
	if !templates.CanBeEdited(filename) {
		return http.StatusNotAcceptable, errors.New("File format not supported.")
	}

	// Check if the file exists.
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return http.StatusNotFound, nil
	} else if os.IsPermission(err) {
		return http.StatusForbidden, nil
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	// Open the file and check if there was some error while opening
	file, err := ioutil.ReadFile(filename)
	if os.IsPermission(err) {
		return http.StatusForbidden, nil
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	// Create a new editor variable and set the extension
	page := new(editor)
	page.Mode = strings.TrimPrefix(filepath.Ext(filename), ".")
	page.Name = strings.Replace(filename, c.Path, "", 1)
	page.Config = c
	page.IsPost = false

	// Sanitize the extension
	page.Mode = sanitizeMode(page.Mode)

	// Handle the content depending on the file extension
	switch page.Mode {
	case "markdown", "asciidoc", "rst":
		if hasFrontMatterRune(file) {
			// Starts a new buffer and parses the file using Hugo's functions
			buffer := bytes.NewBuffer(file)
			file, err := parser.ReadFrom(buffer)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if strings.Contains(string(file.FrontMatter()), "date") {
				page.IsPost = true
			}

			// Parses the page content and the frontmatter
			page.Content = strings.TrimSpace(string(file.Content()))
			page.FrontMatter, page.Name, err = frontmatter.Pretty(file.FrontMatter())
			page.Class = "complete"
		} else {
			// The editor will handle only content
			page.Class = "content-only"
			page.Content = string(file)
		}
	case "json", "toml", "yaml":
		// Defines the class and declares an error
		page.Class = "frontmatter-only"
		var err error

		// Checks if the file already has the frontmatter rune and parses it
		if hasFrontMatterRune(file) {
			page.FrontMatter, _, err = frontmatter.Pretty(file)
		} else {
			page.FrontMatter, _, err = frontmatter.Pretty(appendFrontMatterRune(file, page.Mode))
		}

		// Check if there were any errors
		if err != nil {
			return http.StatusInternalServerError, err
		}
	default:
		// The editor will handle only content
		page.Class = "content-only"
		page.Content = string(file)
	}

	// Create the functions map, then the template, check for erros and
	// execute the template if there aren't errors
	functions := template.FuncMap{
		"SplitCapitalize": templates.SplitCapitalize,
		"Defined":         templates.Defined,
	}

	tpl, err := templates.Get(r, functions, "editor", "frontmatter")

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, tpl.Execute(w, page)
}

func hasFrontMatterRune(file []byte) bool {
	return strings.HasPrefix(string(file), "---") ||
		strings.HasPrefix(string(file), "+++") ||
		strings.HasPrefix(string(file), "{")
}

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

func sanitizeMode(extension string) string {
	switch extension {
	case "md", "markdown", "mdown", "mmark":
		return "markdown"
	case "asciidoc", "adoc", "ad":
		return "asciidoc"
	case "rst":
		return "rst"
	case "html", "htm":
		return "html"
	case "js":
		return "javascript"
	default:
		return extension
	}
}

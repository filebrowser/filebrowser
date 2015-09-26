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
	"github.com/hacdias/caddy-hugo/frontmatter"
	"github.com/hacdias/caddy-hugo/utils"
	"github.com/spf13/hugo/parser"
)

// GET handles the GET method on editor page
func GET(w http.ResponseWriter, r *http.Request, c *config.Config, filename string) (int, error) {
	// Check if the file format is supported. If not, send a "Not Acceptable"
	// header and an error
	if !utils.CanBeEdited(filename) {
		return http.StatusNotAcceptable, errors.New("File format not supported.")
	}

	// Check if the file exists. If it doesn't, send a "Not Found" message
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return http.StatusNotFound, nil
	}

	// Open the file and check if there was some error while opening
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Create a new editor variable and set the extension
	page := new(editor)
	page.Mode = strings.TrimPrefix(filepath.Ext(filename), ".")
	page.Name = filename
	page.Config = c
	page.IsPost = false

	// Sanitize the extension
	page.Mode = sanitizeMode(page.Mode)

	// Handle the content depending on the file extension
	switch page.Mode {
	case "markdown":
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
			page.FrontMatter, err = frontmatter.Pretty(file.FrontMatter())
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
			page.FrontMatter, err = frontmatter.Pretty(file)
		} else {
			page.FrontMatter, err = frontmatter.Pretty(appendFrontMatterRune(file, page.Mode))
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
		"SplitCapitalize": utils.SplitCapitalize,
		"Defined":         utils.Defined,
	}

	tpl, err := utils.GetTemplate(r, functions, "editor", "frontmatter")

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
	case "markdown", "md":
		return "markdown"
	case "css", "scss":
		return "css"
	case "html":
		return "htmlmixed"
	case "js":
		return "javascript"
	default:
		return extension
	}
}

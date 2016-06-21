package hugo

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager"
	"github.com/spf13/hugo/parser"
)

type editor struct {
	Name        string
	Class       string
	IsPost      bool
	Mode        string
	Content     string
	FrontMatter interface{}
}

// GET handles the GET method on editor page
func (h Hugo) GET(w http.ResponseWriter, r *http.Request, filename string) (int, error) {
	// Check if the file exists.
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return http.StatusNotFound, err
	} else if os.IsPermission(err) {
		return http.StatusForbidden, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	// Open the file and check if there was some error while opening
	file, err := ioutil.ReadFile(filename)
	if os.IsPermission(err) {
		return http.StatusForbidden, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	page := &filemanager.Page{
		Info: &filemanager.PageInfo{
			IsDir:  false,
			Config: &h.FileManager.Configs[0],
			Name:   strings.Replace(filename, h.Config.Root, "", 1),
		},
	}

	// Create a new editor variable and set the extension
	data := new(editor)
	data.Mode = strings.TrimPrefix(filepath.Ext(filename), ".")
	data.Name = strings.Replace(filename, h.Config.Root, "", 1)
	data.IsPost = false
	data.Mode = sanitizeMode(data.Mode)

	var parserPage parser.Page

	// Handle the content depending on the file extension
	switch data.Mode {
	case "markdown", "asciidoc", "rst":
		if hasFrontMatterRune(file) {
			// Starts a new buffer and parses the file using Hugo's functions
			buffer := bytes.NewBuffer(file)
			parserPage, err = parser.ReadFrom(buffer)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if strings.Contains(string(parserPage.FrontMatter()), "date") {
				data.IsPost = true
			}

			// Parses the page content and the frontmatter
			data.Content = strings.TrimSpace(string(parserPage.Content()))
			data.FrontMatter, data.Name, err = Pretty(parserPage.FrontMatter())
			data.Class = "complete"
		} else {
			// The editor will handle only content
			data.Class = "content-only"
			data.Content = string(file)
		}
	case "json", "toml", "yaml":
		// Defines the class and declares an error
		data.Class = "frontmatter-only"

		// Checks if the file already has the frontmatter rune and parses it
		if hasFrontMatterRune(file) {
			data.FrontMatter, _, err = Pretty(file)
		} else {
			data.FrontMatter, _, err = Pretty(appendFrontMatterRune(file, data.Mode))
		}

		// Check if there were any errors
		if err != nil {
			return http.StatusInternalServerError, err
		}
	default:
		// The editor will handle only content
		data.Class = "content-only"
		data.Content = string(file)
	}

	// Create the functions map, then the template, check for erros and
	// execute the template if there aren't errors
	functions := template.FuncMap{
		"SplitCapitalize": SplitCapitalize,
		"Defined":         Defined,
	}

	var code int

	page.Info.Data = data

	templates := []string{"frontmatter", "editor"}
	for _, t := range templates {
		code, err = page.AddTemplate(t, Asset, functions)
		if err != nil {
			return code, err
		}
	}

	templates = []string{"actions", "base"}
	for _, t := range templates {
		code, err = page.AddTemplate(t, filemanager.Asset, nil)
		if err != nil {
			return code, err
		}
	}

	return page.PrintAsHTML(w)
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

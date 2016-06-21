package templates

import (
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/hacdias/caddy-hugo/routes/assets"
)

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

// Get is used to get a ready to use template based on the url and on
// other sent templates
func Get(r *http.Request, functions template.FuncMap, templates ...string) (*template.Template, error) {
	// If this is a pjax request, use the minimal template to send only
	// the main content
	if r.Header.Get("X-PJAX") == "true" {
		templates = append(templates, "base_minimal")
	} else {
		templates = append(templates, "base_full")
	}

	var tpl *template.Template

	// For each template, add it to the the tpl variable
	for i, t := range templates {
		// Get the template from the assets
		page, err := assets.Asset("templates/" + t + ".tmpl")

		// Check if there is some error. If so, the template doesn't exist
		if err != nil {
			log.Print(err)
			return new(template.Template), err
		}

		// If it's the first iteration, creates a new template and add the
		// functions map
		if i == 0 {
			tpl, err = template.New(t).Funcs(functions).Parse(string(page))
		} else {
			tpl, err = tpl.Parse(string(page))
		}

		if err != nil {
			log.Print(err)
			return new(template.Template), err
		}
	}

	return tpl, nil
}

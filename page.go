package filemanager

import (
	"bytes"
	"html/template"
	"log"
	"strings"
)

type Page struct {
	Name   string
	Path   string
	Config *Config
	Data   interface{}
}

func (f FileManager) formatAsHTML(page *Page, templates ...string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	templates = append(templates, "base")
	var tpl *template.Template

	// For each template, add it to the the tpl variable
	for i, t := range templates {
		// Get the template from the assets
		page, err := Asset("templates/" + t + ".tmpl")

		// Check if there is some error. If so, the template doesn't exist
		if err != nil {
			log.Print(err)
			return new(bytes.Buffer), err
		}

		// If it's the first iteration, creates a new template and add the
		// functions map
		if i == 0 {
			tpl, err = template.New(t).Parse(string(page))
		} else {
			tpl, err = tpl.Parse(string(page))
		}

		if err != nil {
			log.Print(err)
			return new(bytes.Buffer), err
		}
	}

	err := tpl.Execute(buf, page)
	return buf, err
}

// BreadcrumbMap returns p.Path where every element is a map
// of URLs and path segment names.
func (p Page) BreadcrumbMap() map[string]string {
	result := map[string]string{}

	if len(p.Path) == 0 {
		return result
	}

	// skip trailing slash
	lpath := p.Path
	if lpath[len(lpath)-1] == '/' {
		lpath = lpath[:len(lpath)-1]
	}

	parts := strings.Split(lpath, "/")
	for i, part := range parts {
		if i == 0 && part == "" {
			// Leading slash (root)
			result["/"] = "/"
			continue
		}
		result[strings.Join(parts[:i+1], "/")] = part
	}

	return result
}

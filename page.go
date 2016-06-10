package filemanager

import (
	"bytes"
	"html/template"
	"log"
)

type Page struct {
	Config *Config
	Data   interface{}
}

func (f FileManager) formatAsHTML(data interface{}, fmc *Config, templates ...string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	pg := &Page{
		Config: fmc,
		Data:   data,
	}

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

	err := Template.Execute(buf, pg)
	return buf, err
}

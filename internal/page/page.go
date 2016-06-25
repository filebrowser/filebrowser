package page

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/hacdias/caddy-filemanager/internal/assets"
	"github.com/hacdias/caddy-filemanager/internal/config"
	"github.com/hacdias/caddy-filemanager/utils/variables"
)

// Page contains the informations and functions needed to show the page
type Page struct {
	*Info
}

// Info contains the information of a page
type Info struct {
	Name   string
	Path   string
	IsDir  bool
	Config *config.Config
	Data   interface{}
}

// BreadcrumbMap returns p.Path where every element is a map
// of URLs and path segment names.
func (i Info) BreadcrumbMap() map[string]string {
	result := map[string]string{}

	if len(i.Path) == 0 {
		return result
	}

	// skip trailing slash
	lpath := i.Path
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

// PreviousLink returns the path of the previous folder
func (i Info) PreviousLink() string {
	parts := strings.Split(strings.TrimSuffix(i.Path, "/"), "/")
	if len(parts) <= 1 {
		return ""
	}

	if parts[len(parts)-2] == "" {
		if i.Config.BaseURL == "" {
			return "/"
		}
		return i.Config.BaseURL
	}

	return parts[len(parts)-2]
}

// PrintAsHTML formats the page in HTML and executes the template
func (p Page) PrintAsHTML(w http.ResponseWriter, templates ...string) (int, error) {
	// Create the functions map, then the template, check for erros and
	// execute the template if there aren't errors
	functions := template.FuncMap{
		"SplitCapitalize": variables.SplitCapitalize,
		"Defined":         variables.Defined,
	}

	templates = append(templates, "actions", "base")
	var tpl *template.Template

	// For each template, add it to the the tpl variable
	for i, t := range templates {
		// Get the template from the assets
		page, err := assets.Asset("templates/" + t + ".tmpl")

		// Check if there is some error. If so, the template doesn't exist
		if err != nil {
			log.Print(err)
			return http.StatusInternalServerError, err
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
			return http.StatusInternalServerError, err
		}
	}

	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, p.Info)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = buf.WriteTo(w)
	return http.StatusOK, nil
}

// PrintAsJSON prints the current page infromation in JSON
func (p Page) PrintAsJSON(w http.ResponseWriter) (int, error) {
	marsh, err := json.Marshal(p.Info.Data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return w.Write(marsh)
}

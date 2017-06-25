package filemanager

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"github.com/hacdias/filemanager/variables"
)

// functions contains the non-standard functions that are available
// to use on the HTML templates.
var functions = template.FuncMap{
	"Defined": variables.FieldInStruct,
	"CSS": func(s string) template.CSS {
		return template.CSS(s)
	},
	"Marshal": func(v interface{}) template.JS {
		a, _ := json.Marshal(v)
		return template.JS(a)
	},
	"EncodeBase64": func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	},
}

// page contains the information needed to fill a page template.
type page struct {
	minimal   bool
	Name      string
	Path      string
	IsDir     bool
	User      *User
	PrefixURL string
	BaseURL   string
	WebDavURL string
	Data      interface{}
	Editor    bool
	Display   string
}

// breadcrumbItem contains the Name and the URL of a breadcrumb piece.
type breadcrumbItem struct {
	Name string
	URL  string
}

// BreadcrumbMap returns p.Path where every element is a map
// of URLs and path segment names.
func (p page) BreadcrumbMap() []breadcrumbItem {
	result := []breadcrumbItem{}

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
		if i == len(parts)-1 {
			continue
		}

		if i == 0 && part == "" {
			result = append([]breadcrumbItem{{
				Name: "/",
				URL:  "/",
			}}, result...)
			continue
		}

		result = append([]breadcrumbItem{{
			Name: part,
			URL:  strings.Join(parts[:i+1], "/") + "/",
		}}, result...)
	}

	return result
}

// PreviousLink returns the URL of the previous folder.
func (p page) PreviousLink() string {
	path := strings.TrimSuffix(p.Path, "/")
	path = strings.TrimPrefix(path, "/")
	path = p.BaseURL + "/" + path
	path = path[0 : len(path)-len(p.Name)]

	if len(path) < len(p.BaseURL+"/") {
		return ""
	}

	return path
}

// PrintAsHTML formats the page in HTML and executes the template
func (p page) PrintAsHTML(w http.ResponseWriter, box *rice.Box, templates ...string) (int, error) {
	if p.minimal {
		templates = append(templates, "minimal")
	} else {
		templates = append(templates, "base")
	}

	var tpl *template.Template

	// For each template, add it to the the tpl variable
	for i, t := range templates {
		// Get the template from the assets
		Page, err := box.String(t + ".tmpl")

		// Check if there is some error. If so, the template doesn't exist
		if err != nil {
			log.Print(err)
			return http.StatusInternalServerError, err
		}

		// If it's the first iteration, creates a new template and add the
		// functions map
		if i == 0 {
			tpl, err = template.New(t).Funcs(functions).Parse(Page)
		} else {
			tpl, err = tpl.Parse(string(Page))
		}

		if err != nil {
			log.Print(err)
			return http.StatusInternalServerError, err
		}
	}

	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, p)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = buf.WriteTo(w)
	return http.StatusOK, err
}

// PrintAsJSON prints the current Page information in JSON
func (p page) PrintAsJSON(w http.ResponseWriter) (int, error) {
	marsh, err := json.MarshalIndent(p.Data, "", "    ")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(marsh); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

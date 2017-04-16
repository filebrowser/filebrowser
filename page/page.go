// Package page is used to render the HTML to the end user
package page

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/hacdias/caddy-filemanager/assets"
	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/utils/variables"
)

// Page contains the informations and functions needed to show the Page
type Page struct {
	*Info
	Minimal bool
}

// Info contains the information of a Page
type Info struct {
	Name    string
	Path    string
	IsDir   bool
	User    *config.User
	Config  *config.Config
	Data    interface{}
	Editor  bool
	Display string
	Token   string
}

// BreadcrumbMapItem ...
type BreadcrumbMapItem struct {
	Name string
	URL  string
}

// BreadcrumbMap returns p.Path where every element is a map
// of URLs and path segment names.
func (i Info) BreadcrumbMap() []BreadcrumbMapItem {
	result := []BreadcrumbMapItem{}

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
		if i == len(parts)-1 {
			continue
		}

		if i == 0 && part == "" {
			result = append([]BreadcrumbMapItem{{
				Name: "/",
				URL:  "/",
			}}, result...)
			continue
		}

		result = append([]BreadcrumbMapItem{{
			Name: part,
			URL:  strings.Join(parts[:i+1], "/") + "/",
		}}, result...)
	}

	return result
}

// PreviousLink returns the path of the previous folder
func (i Info) PreviousLink() string {
	path := strings.TrimSuffix(i.Path, "/")
	path = strings.TrimPrefix(path, "/")
	path = i.Config.AbsoluteURL() + "/" + path
	path = path[0 : len(path)-len(i.Name)]

	if len(path) < len(i.Config.AbsoluteURL()+"/") {
		return ""
	}

	return path
}

// PrintAsHTML formats the page in HTML and executes the template
func (p Page) PrintAsHTML(w http.ResponseWriter, templates ...string) (int, error) {
	// Create the functions map, then the template, check for erros and
	// execute the template if there aren't errors
	functions := template.FuncMap{
		"Defined": variables.Defined,
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

	if p.Minimal {
		templates = append(templates, "minimal")
	} else {
		templates = append(templates, "base")
	}

	var tpl *template.Template

	// For each template, add it to the the tpl variable
	for i, t := range templates {
		// Get the template from the assets
		Page, err := assets.Asset("templates/" + t + ".tmpl")

		// Check if there is some error. If so, the template doesn't exist
		if err != nil {
			log.Print(err)
			return http.StatusInternalServerError, err
		}

		// If it's the first iteration, creates a new template and add the
		// functions map
		if i == 0 {
			tpl, err = template.New(t).Funcs(functions).Parse(string(Page))
		} else {
			tpl, err = tpl.Parse(string(Page))
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
	return http.StatusOK, err
}

// PrintAsJSON prints the current Page information in JSON
func (p Page) PrintAsJSON(w http.ResponseWriter) (int, error) {
	marsh, err := json.MarshalIndent(p.Info.Data, "", "    ")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(marsh); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

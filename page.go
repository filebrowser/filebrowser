package filemanager

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// Page contains the informations and functions needed to show the page
type Page struct {
	Info *PageInfo
	Tpl  *template.Template
}

// AssetFunc is an Assets function
type AssetFunc func(name string) ([]byte, error)

// PageInfo contains the information of a page
type PageInfo struct {
	Name   string
	Path   string
	IsDir  bool
	Config *Config
	Data   interface{}
}

// BreadcrumbMap returns p.Path where every element is a map
// of URLs and path segment names.
func (p PageInfo) BreadcrumbMap() map[string]string {
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

// PreviousLink returns the path of the previous folder
func (p PageInfo) PreviousLink() string {
	parts := strings.Split(strings.TrimSuffix(p.Path, "/"), "/")
	if len(parts) <= 1 {
		return ""
	}

	if parts[len(parts)-2] == "" {
		if p.Config.BaseURL == "" {
			return "/"
		}
		return p.Config.BaseURL
	}

	return parts[len(parts)-2]
}

// AddTemplate adds a template file to the page template
func (p *Page) AddTemplate(name string, assets AssetFunc, functions template.FuncMap) (int, error) {

	// Get the template from the assets
	page, err := assets("templates/" + name + ".tmpl")

	// Check if there is some error. If so, the template doesn't exist
	if err != nil {
		log.Print(err)
		return http.StatusInternalServerError, err
	}

	// If it's the first iteration, creates a new template and add the
	// functions map
	if p.Tpl == nil {
		p.Tpl, err = template.New(name).Funcs(functions).Parse(string(page))
	} else {
		p.Tpl, err = p.Tpl.Parse(string(page))
	}

	if err != nil {
		log.Print(err)
		return http.StatusInternalServerError, err
	}
	return 0, nil
}

// PrintAsHTML formats the page in HTML and executes the template
func (p Page) PrintAsHTML(w http.ResponseWriter) (int, error) {
	buf := &bytes.Buffer{}
	err := p.Tpl.Execute(buf, p.Info)

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

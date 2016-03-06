package templates

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"strings"
	"unicode"

	"github.com/hacdias/caddy-hugo/assets"
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

// Defined checks if variable is defined in a struct
func Defined(data interface{}, field string) bool {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()

	if t.Kind() != reflect.Struct {
		log.Print("Non-struct type not allowed.")
		return false
	}

	_, b := t.FieldByName(field)
	return b
}

// Dict allows to send more than one variable into a template
func Dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}

	return dict, nil
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

var splitCapitalizeExceptions = map[string]string{
	"youtube":    "YouTube",
	"github":     "GitHub",
	"googleplus": "Google Plus",
	"linkedin":   "LinkedIn",
}

// SplitCapitalize splits a string by its uppercase letters and capitalize the
// first letter of the string
func SplitCapitalize(name string) string {
	if val, ok := splitCapitalizeExceptions[strings.ToLower(name)]; ok {
		return val
	}

	var words []string
	l := 0
	for s := name; s != ""; s = s[l:] {
		l = strings.IndexFunc(s[1:], unicode.IsUpper) + 1
		if l <= 0 {
			l = len(s)
		}
		words = append(words, s[:l])
	}

	name = ""

	for _, element := range words {
		name += element + " "
	}

	name = strings.ToLower(name[:len(name)-1])
	name = strings.ToUpper(string(name[0])) + name[1:]

	return name
}

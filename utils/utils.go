package utils

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"text/template"
	"unicode"

	"github.com/hacdias/caddy-hugo/assets"
	"github.com/hacdias/caddy-hugo/config"
	"github.com/spf13/hugo/commands"
	"github.com/spf13/viper"
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

// CopyFile is used to copy a file
func CopyFile(old, new string) error {
	// Open the file and create a new one
	r, err := os.Open(old)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(new)
	if err != nil {
		return err
	}
	defer w.Close()

	// Copy the content
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}

	return nil
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

// GetTemplate is used to get a ready to use template based on the url and on
// other sent templates
func GetTemplate(r *http.Request, functions template.FuncMap, templates ...string) (*template.Template, error) {
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

// IsMap checks if some variable is a map
func IsMap(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Map
}

// IsSlice checks if some variable is a slice
func IsSlice(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Slice
}

// ParseComponents parses the components of an URL creating an array
func ParseComponents(r *http.Request) []string {
	//The URL that the user queried.
	path := r.URL.Path
	path = strings.TrimSpace(path)
	//Cut off the leading and trailing forward slashes, if they exist.
	//This cuts off the leading forward slash.
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	//This cuts off the trailing forward slash.
	if strings.HasSuffix(path, "/") {
		cutOffLastCharLen := len(path) - 1
		path = path[:cutOffLastCharLen]
	}
	//We need to isolate the individual components of the path.
	components := strings.Split(path, "/")
	return components
}

// Run is used to run the static website generator
func Run(c *config.Config) {
	os.RemoveAll(c.Path + "public")

	commands.MainSite = nil
	viper.Reset()
	commands.HugoCmd.ParseFlags(c.Args)
	if err := commands.HugoCmd.RunE(nil, nil); err != nil {
		log.Print(err)
	}
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

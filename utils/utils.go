package utils

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
	"unicode"
)

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

// IsMap checks if some variable is a map
func IsMap(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Map
}

// IsSlice checks if some variable is a slice
func IsSlice(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Slice
}

// IsArray checks if some variable is an array
func IsArray(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Array
}

// IsString checks if some variable is a string
func IsString(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.String
}

// IsInt checks if some variable is an integer
func IsInt(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Int
}

// IsBool checks if some variable is a boolean
func IsBool(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Bool
}

// IsInterface checks if some variable is an interface
func IsInterface(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Interface
}

// IsMarkdownFile checks if a filename belongs to a markdown file
func IsMarkdownFile(filename string) bool {
	return strings.HasSuffix(filename, ".markdown") || strings.HasSuffix(filename, ".md")
}

// SplitCapitalize splits a string by its uppercase letters and capitalize the
// first letter of the string
func SplitCapitalize(name string) string {
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
	name = strings.ToUpper(string(name[0])) + name[1:len(name)]

	return name
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

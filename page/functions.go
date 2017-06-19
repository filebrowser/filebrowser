package page

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"log"
	"reflect"
)

// Create the functions map, then the template, check for erros and
// execute the template if there aren't errors
var functionMap = template.FuncMap{
	"Defined":      defined,
	"CSS":          css,
	"Marshal":      marshal,
	"EncodeBase64": encodeBase64,
}

// defined checks if variable is defined in a struct
func defined(data interface{}, field string) bool {
	t := reflect.Indirect(reflect.ValueOf(data)).Type()

	if t.Kind() != reflect.Struct {
		log.Print("Non-struct type not allowed.")
		return false
	}

	_, b := t.FieldByName(field)
	return b
}

// css returns the sanitized and safe css
func css(s string) template.CSS {
	return template.CSS(s)
}

// marshal converts an interface to json and sanitizes it
func marshal(v interface{}) template.JS {
	a, _ := json.Marshal(v)
	return template.JS(a)
}

// encodeBase64 encodes a string in base 64
func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

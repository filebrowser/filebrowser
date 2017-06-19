package filemanager

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"reflect"
)

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

// errorToHTTPCode converts errors to HTTP Status Code.
func errorToHTTPCode(err error, gone bool) int {
	switch {
	case os.IsPermission(err):
		return http.StatusForbidden
	case os.IsNotExist(err):
		if !gone {
			return http.StatusNotFound
		}

		return http.StatusGone
	case os.IsExist(err):
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}

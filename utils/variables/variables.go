package variables

import (
	"errors"
	"log"
	"reflect"
)

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

// StringInSlice checks if a slice contains a string
func StringInSlice(a string, list []string) (bool, int) {
	for i, b := range list {
		if b == a {
			return true, i
		}
	}
	return false, 0
}

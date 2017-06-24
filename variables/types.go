package variables

import "reflect"

// IsMap checks if some variable is a map
func IsMap(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Map
}

// IsSlice checks if some variable is a slice
func IsSlice(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Slice
}

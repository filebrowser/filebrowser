package frontmatter

import "reflect"

// isMap checks if some variable is a map
func isMap(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Map
}

// isSlice checks if some variable is a slice
func isSlice(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Slice
}

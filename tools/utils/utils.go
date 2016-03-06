package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

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

func RespondJSON(w http.ResponseWriter, message map[string]string, code int, err error) (int, error) {
	msg, msgErr := json.Marshal(message)

	if msgErr != nil {
		return 500, msgErr
	}

	if code == 500 && err != nil {
		err = errors.New(message["message"])
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(msg)
	return 0, err
}

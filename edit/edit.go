package edit

import (
	"io/ioutil"
	"net/http"
	"os"
)

type Page struct {
}

// Execute sth
func Execute(w http.ResponseWriter, r *http.Request, file string) (int, error) {
	if r.Method == "POST" {
		// it's saving the post
	} else {
		// check if the file exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return 404, nil
		}

		file, _ := ioutil.ReadFile(file)
		w.Write([]byte(string(file)))
	}

	return 200, nil
}

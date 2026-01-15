package fbhttp

import (
	"net/http"
	"sort"

	"github.com/filebrowser/filebrowser/v2/files"
)

var encodingsHandler = withUser(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
	list := make([]string, 0, len(files.Encodings))
	for name := range files.Encodings {
		list = append(list, name)
	}

	sort.Strings(list)

	return renderJSON(w, r, map[string]interface{}{
		"encodings": list,
	})
})

package file

import (
	"net/http"

	"github.com/hacdias/caddy-filemanager/internal/config"
)

// Update is
func (i *Info) Update(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {

	/*
	   // POST handles the POST method on editor page
	   func POST(w http.ResponseWriter, r *http.Request, c *config.Config, filename string) (int, error) {
	   	var data info

	   	// Get the JSON information sent using a buffer
	   	rawBuffer := new(bytes.Buffer)
	   	rawBuffer.ReadFrom(r.Body)
	   	err := json.Unmarshal(rawBuffer.Bytes(), &data)

	   	fmt.Println(string(rawBuffer.Bytes()))

	   	if err != nil {
	   		return server.RespondJSON(w, &response{"Error decrypting json."}, http.StatusInternalServerError, err)
	   	}

	   	// Initializes the file content to write
	   	var file []byte
	   	var code int

	   	switch data.ContentType {
	   	case "frontmatter-only":
	   		file, code, err = parseFrontMatterOnlyFile(data, filename)
	   		if err != nil {
	   			return server.RespondJSON(w, &response{err.Error()}, code, err)
	   		}
	   	case "content-only":
	   		// The main content of the file
	   		mainContent := data.Content["content"].(string)
	   		mainContent = strings.TrimSpace(mainContent)

	   		file = []byte(mainContent)
	   	case "complete":
	   		file, code, err = parseCompleteFile(data, filename, c)
	   		if err != nil {
	   			return server.RespondJSON(w, &response{err.Error()}, code, err)
	   		}
	   	default:
	   		return server.RespondJSON(w, &response{"Invalid content type."}, http.StatusBadRequest, nil)
	   	}

	   	// Write the file
	   	err = ioutil.WriteFile(filename, file, 0666)

	   	if err != nil {
	   		return server.RespondJSON(w, &response{err.Error()}, http.StatusInternalServerError, err)
	   	}

	   	if data.Regenerate {
	   		go hugo.Run(c, false)
	   	}

	   	return server.RespondJSON(w, nil, http.StatusOK, nil)
	   }
	*/
	return 0, nil
}

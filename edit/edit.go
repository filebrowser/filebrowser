package edit

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hacdias/caddy-hugo/frontmatter"
	"github.com/hacdias/caddy-hugo/page"
	"github.com/spf13/hugo/parser"
)

type information struct {
	Name        string
	Content     string
	FrontMatter interface{}
}

// Execute sth
func Execute(w http.ResponseWriter, r *http.Request) (int, error) {
	filename := strings.Replace(r.URL.Path, "/admin/edit/", "", 1)

	if r.Method == "POST" {
		// Get the JSON information sent using a buffer
		rawBuffer := new(bytes.Buffer)
		rawBuffer.ReadFrom(r.Body)

		// Creates the raw file "map" using the JSON
		var rawFile map[string]interface{}
		json.Unmarshal(rawBuffer.Bytes(), &rawFile)

		// The main content of the file
		mainContent := rawFile["content"].(string)
		mainContent = "\n\n" + strings.TrimSpace(mainContent)

		// Removes the main content from the rest of the frontmatter
		delete(rawFile, "content")

		// Converts the frontmatter in JSON
		jsonFrontmatter, err := json.Marshal(rawFile)

		if err != nil {
			log.Print(err)
			return 500, err
		}

		// Indents the json
		frontMatterBuffer := new(bytes.Buffer)
		json.Indent(frontMatterBuffer, jsonFrontmatter, "", "  ")

		// Generates the final file
		file := new(bytes.Buffer)
		file.Write(frontMatterBuffer.Bytes())
		file.Write([]byte(mainContent))

		err = ioutil.WriteFile(filename, file.Bytes(), 0666)

		if err != nil {
			log.Print(err)
			return 500, err
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	} else {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.Print(err)
			return 404, nil
		}

		reader, err := os.Open(filename)

		if err != nil {
			log.Print(err)
			return 500, err
		}

		file, err := parser.ReadFrom(reader)

		inf := new(information)
		inf.Content = strings.TrimSpace(string(file.Content()))
		inf.FrontMatter, err = frontmatter.Pretty(file.FrontMatter())

		if err != nil {
			log.Print(err)
			return 500, err
		}

		page := new(page.Page)
		page.Name = "Edit"
		page.Body = inf
		return page.Render(w, r, "edit", "frontmatter")
	}

	return 200, nil
}

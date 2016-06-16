package editor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/hacdias/caddy-hugo/tools/hugo"
	"github.com/hacdias/caddy-hugo/tools/server"
	"github.com/robfig/cron"
	"github.com/spf13/cast"
	"github.com/spf13/hugo/parser"
)

type info struct {
	ContentType string
	Schedule    bool
	Regenerate  bool
	Content     map[string]interface{}
}

type response struct {
	Message string `json:"message"`
}

// POST handles the POST method on editor page
func POST(w http.ResponseWriter, r *http.Request) (int, error) {
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
		file, code, err = parseFrontMatterOnlyFile(data)
		if err != nil {
			return server.RespondJSON(w, &response{err.Error()}, code, err)
		}
	case "content-only":
		// The main content of the file
		mainContent := data.Content["content"].(string)
		mainContent = strings.TrimSpace(mainContent)

		file = []byte(mainContent)
	case "complete":
		file, code, err = parseCompleteFile(data)
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
		go hugo.Run(conf, false)
	}

	return server.RespondJSON(w, nil, http.StatusOK, nil)
}

func parseFrontMatterOnlyFile(data info) ([]byte, int, error) {
	frontmatter := strings.TrimPrefix(filepath.Ext(filename), ".")
	var mark rune

	switch frontmatter {
	case "toml":
		mark = rune('+')
	case "json":
		mark = rune('{')
	case "yaml":
		mark = rune('-')
	default:
		return []byte{}, http.StatusBadRequest, errors.New("Can't define the frontmatter.")
	}

	f, err := parser.InterfaceToFrontMatter(data.Content, mark)
	fString := string(f)

	// If it's toml or yaml, strip frontmatter identifier
	if frontmatter == "toml" {
		fString = strings.TrimSuffix(fString, "+++\n")
		fString = strings.TrimPrefix(fString, "+++\n")
	}

	if frontmatter == "yaml" {
		fString = strings.TrimSuffix(fString, "---\n")
		fString = strings.TrimPrefix(fString, "---\n")
	}

	f = []byte(fString)

	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}

	return f, http.StatusOK, nil
}

func parseCompleteFile(data info) ([]byte, int, error) {
	// The main content of the file
	mainContent := data.Content["content"].(string)
	mainContent = "\n\n" + strings.TrimSpace(mainContent) + "\n"

	// Removes the main content from the rest of the frontmatter
	delete(data.Content, "content")

	if _, ok := data.Content["date"]; ok {
		data.Content["date"] = data.Content["date"].(string) + ":00"
	}

	// Schedule the post
	if data.Schedule {
		t := cast.ToTime(data.Content["date"])

		scheduler := cron.New()
		scheduler.AddFunc(t.In(time.Now().Location()).Format("05 04 15 02 01 *"), func() {
			// Set draft to false
			data.Content["draft"] = false

			// Converts the frontmatter in JSON
			jsonFrontmatter, err := json.Marshal(data.Content)

			if err != nil {
				return
			}

			// Indents the json
			frontMatterBuffer := new(bytes.Buffer)
			json.Indent(frontMatterBuffer, jsonFrontmatter, "", "  ")

			// Generates the final file
			f := new(bytes.Buffer)
			f.Write(frontMatterBuffer.Bytes())
			f.Write([]byte(mainContent))
			file := f.Bytes()

			// Write the file
			if err = ioutil.WriteFile(filename, file, 0666); err != nil {
				return
			}

			go hugo.Run(conf, false)
		})
		scheduler.Start()
	}

	// Converts the frontmatter in JSON
	jsonFrontmatter, err := json.Marshal(data.Content)

	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}

	// Indents the json
	frontMatterBuffer := new(bytes.Buffer)
	json.Indent(frontMatterBuffer, jsonFrontmatter, "", "  ")

	// Generates the final file
	f := new(bytes.Buffer)
	f.Write(frontMatterBuffer.Bytes())
	f.Write([]byte(mainContent))
	return f.Bytes(), http.StatusOK, nil
}

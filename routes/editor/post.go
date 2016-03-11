package editor

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/tools/hugo"
	"github.com/hacdias/caddy-hugo/tools/server"
	"github.com/robfig/cron"
	"github.com/spf13/cast"
	"github.com/spf13/hugo/parser"
)

var schedule, contentType, regenerate string

// POST handles the POST method on editor page
func POST(w http.ResponseWriter, r *http.Request, c *config.Config, filename string) (int, error) {
	// Get the JSON information sent using a buffer
	rawBuffer := new(bytes.Buffer)
	rawBuffer.ReadFrom(r.Body)

	// Creates the data map using the JSON
	var data map[string]interface{}
	json.Unmarshal(rawBuffer.Bytes(), &data)

	// Checks if all the all data is defined
	if _, ok := data["type"]; !ok {
		return server.RespondJSON(w, map[string]string{
			"message": "Content type not set.",
		}, http.StatusBadRequest, nil)
	}

	if _, ok := data["content"]; !ok {
		return server.RespondJSON(w, map[string]string{
			"message": "Content not sent.",
		}, http.StatusBadRequest, nil)
	}

	if _, ok := data["schedule"]; !ok {
		return server.RespondJSON(w, map[string]string{
			"message": "Schedule information not sent.",
		}, http.StatusBadRequest, nil)
	}

	if _, ok := data["regenerate"]; !ok {
		return server.RespondJSON(w, map[string]string{
			"message": "Regenerate information not sent.",
		}, http.StatusBadRequest, nil)
	}

	rawFile := data["content"].(map[string]interface{})
	contentType = data["type"].(string)
	schedule = data["schedule"].(string)
	regenerate = data["regenerate"].(string)

	// Initializes the file content to write
	var file []byte

	switch contentType {
	case "frontmatter-only":
		f, code, err := parseFrontMatterOnlyFile(rawFile, filename)
		if err != nil {
			return server.RespondJSON(w, map[string]string{
				"message": err.Error(),
			}, code, err)
		}

		file = f
	case "content-only":
		// The main content of the file
		mainContent := rawFile["content"].(string)
		mainContent = strings.TrimSpace(mainContent)

		file = []byte(mainContent)
	case "complete":
		f, code, err := parseCompleteFile(r, c, rawFile, filename)
		if err != nil {
			return server.RespondJSON(w, map[string]string{
				"message": err.Error(),
			}, code, err)
		}

		file = f
	default:
		return server.RespondJSON(w, map[string]string{
			"message": "Invalid content type.",
		}, http.StatusBadRequest, nil)
	}

	// Write the file
	err := ioutil.WriteFile(filename, file, 0666)

	if err != nil {
		return server.RespondJSON(w, map[string]string{
			"message": err.Error(),
		}, http.StatusInternalServerError, err)
	}

	if regenerate == "true" {
		go hugo.Run(c, false)
	}

	return server.RespondJSON(w, map[string]string{}, http.StatusOK, nil)
}

func parseFrontMatterOnlyFile(rawFile map[string]interface{}, filename string) ([]byte, int, error) {
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

	f, err := parser.InterfaceToFrontMatter(rawFile, mark)
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

func parseCompleteFile(r *http.Request, c *config.Config, rawFile map[string]interface{}, filename string) ([]byte, int, error) {
	// The main content of the file
	mainContent := rawFile["content"].(string)
	mainContent = "\n\n" + strings.TrimSpace(mainContent) + "\n"

	// Removes the main content from the rest of the frontmatter
	delete(rawFile, "content")

	// Schedule the post
	if schedule == "true" {
		t := cast.ToTime(rawFile["date"].(string))

		scheduler := cron.New()
		scheduler.AddFunc(t.In(time.Now().Location()).Format("05 04 15 02 01 *"), func() {
			// Set draft to false
			rawFile["draft"] = false

			// Converts the frontmatter in JSON
			jsonFrontmatter, err := json.Marshal(rawFile)

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

			go hugo.Run(c, false)
		})
		scheduler.Start()
	}

	// Converts the frontmatter in JSON
	jsonFrontmatter, err := json.Marshal(rawFile)

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

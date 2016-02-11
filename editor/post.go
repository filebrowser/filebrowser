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
	"github.com/hacdias/caddy-hugo/utils"
	"github.com/robfig/cron"
	"github.com/spf13/cast"
	"github.com/spf13/hugo/parser"
)

// POST handles the POST method on editor page
func POST(w http.ResponseWriter, r *http.Request, c *config.Config, filename string) (int, error) {
	// Get the JSON information sent using a buffer
	rawBuffer := new(bytes.Buffer)
	rawBuffer.ReadFrom(r.Body)

	// Creates the raw file "map" using the JSON
	var rawFile map[string]interface{}
	json.Unmarshal(rawBuffer.Bytes(), &rawFile)

	// Initializes the file content to write
	var file []byte

	switch r.Header.Get("X-Content-Type") {
	case "frontmatter-only":
		f, code, err := parseFrontMatterOnlyFile(rawFile, filename)
		if err != nil {
			w.Write([]byte(err.Error()))
			return code, err
		}

		file = f
	case "content-only":
		// The main content of the file
		mainContent := rawFile["content"].(string)
		mainContent = "\n\n" + strings.TrimSpace(mainContent)

		file = []byte(mainContent)
	case "complete":
		f, code, err := parseCompleteFile(r, c, rawFile, filename)
		if err != nil {
			w.Write([]byte(err.Error()))
			return code, err
		}

		file = f
	default:
		return http.StatusBadRequest, errors.New("X-Content-Type header not defined")
	}

	// Write the file
	err := ioutil.WriteFile(filename, file, 0666)

	if err != nil {
		w.Write([]byte(err.Error()))
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
	return http.StatusOK, nil
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
	if r.Header.Get("X-Schedule") == "true" {
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

			utils.Run(c)
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

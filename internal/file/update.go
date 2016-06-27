package file

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/internal/config"
	"github.com/spf13/hugo/parser"
)

// Update is used to update a file that was edited
func (i *Info) Update(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	var data map[string]interface{}
	kind := r.Header.Get("kind")

	if kind == "" {
		return http.StatusBadRequest, nil
	}

	// Get the JSON information
	rawBuffer := new(bytes.Buffer)
	rawBuffer.ReadFrom(r.Body)
	err := json.Unmarshal(rawBuffer.Bytes(), &data)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	var file []byte
	var code int

	switch kind {
	case "frontmatter-only":
		if file, code, err = parseFrontMatterOnlyFile(data, i.Name); err != nil {
			return http.StatusInternalServerError, err
		}
	case "content-only":
		mainContent := data["content"].(string)
		mainContent = strings.TrimSpace(mainContent)
		file = []byte(mainContent)
	case "complete":
		if file, code, err = parseCompleteFile(data, i.Name); err != nil {
			return http.StatusInternalServerError, err
		}
	default:
		return http.StatusBadRequest, nil
	}

	// Write the file
	err = ioutil.WriteFile(i.Path, file, 0666)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return code, nil
}

func parseFrontMatterOnlyFile(data interface{}, filename string) ([]byte, int, error) {
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

	f, err := parser.InterfaceToFrontMatter(data, mark)
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

func parseCompleteFile(data map[string]interface{}, filename string) ([]byte, int, error) {
	// The main content of the file
	mainContent := data["content"].(string)
	mainContent = "\n\n" + strings.TrimSpace(mainContent) + "\n"

	// Removes the main content from the rest of the frontmatter
	delete(data, "content")

	if _, ok := data["date"]; ok {
		data["date"] = data["date"].(string) + ":00"
	}

	// Converts the frontmatter in JSON
	jsonFrontmatter, err := json.Marshal(data)

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

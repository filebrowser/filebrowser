package directory

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/config"
	"github.com/spf13/hugo/parser"
)

// Update is used to update a file that was edited
func (i *Info) Update(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.User) (int, error) {
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
		if file, code, err = ParseFrontMatterOnlyFile(data, i.Name); err != nil {
			return http.StatusInternalServerError, err
		}
	case "content-only":
		mainContent := data["content"].(string)
		mainContent = strings.TrimSpace(mainContent)
		file = []byte(mainContent)
	case "complete":
		if file, code, err = ParseCompleteFile(data, i.Name, u.FrontMatter); err != nil {
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

// ParseFrontMatterOnlyFile parses a frontmatter only file
func ParseFrontMatterOnlyFile(data interface{}, filename string) ([]byte, int, error) {
	frontmatter := strings.TrimPrefix(filepath.Ext(filename), ".")
	f, code, err := ParseFrontMatter(data, frontmatter)
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
	return f, code, err
}

// ParseFrontMatter is the frontmatter parser
func ParseFrontMatter(data interface{}, frontmatter string) ([]byte, int, error) {
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

	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}

	return f, http.StatusOK, nil
}

// ParseCompleteFile parses a complete file
func ParseCompleteFile(data map[string]interface{}, filename string, frontmatter string) ([]byte, int, error) {
	mainContent := ""

	if _, ok := data["content"]; ok {
		// The main content of the file
		mainContent = data["content"].(string)
		mainContent = "\n\n" + strings.TrimSpace(mainContent) + "\n"

		// Removes the main content from the rest of the frontmatter
		delete(data, "content")
	}

	if _, ok := data["date"]; ok {
		data["date"] = data["date"].(string) + ":00"
	}

	front, code, err := ParseFrontMatter(data, frontmatter)

	if err != nil {
		fmt.Println(frontmatter)
		return []byte{}, code, err
	}

	// Generates the final file
	f := new(bytes.Buffer)
	f.Write(front)
	f.Write([]byte(mainContent))
	return f.Bytes(), http.StatusOK, nil
}

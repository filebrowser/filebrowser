package filemanager

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
	"github.com/hacdias/caddy-filemanager/file"
	"github.com/spf13/hugo/parser"
)

// processPUT is used to update a file that was edited
func processPUT(
	w http.ResponseWriter,
	r *http.Request,
	c *config.Config,
	u *config.User,
	i *file.Info,
) (int, error) {
	var (
		data      map[string]interface{}
		file      []byte
		code      int
		err       error
		kind      string
		rawBuffer = new(bytes.Buffer)
	)

	kind = r.Header.Get("kind")
	rawBuffer.ReadFrom(r.Body)

	if kind != "" {
		err = json.Unmarshal(rawBuffer.Bytes(), &data)

		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	switch kind {
	case "frontmatter-only":
		if file, code, err = parseFrontMatterOnlyFile(data, i.Name()); err != nil {
			return http.StatusInternalServerError, err
		}
	case "content-only":
		mainContent := data["content"].(string)
		mainContent = strings.TrimSpace(mainContent)
		file = []byte(mainContent)
	case "complete":
		if file, code, err = parseCompleteFile(data, i.Name(), u.FrontMatter); err != nil {
			return http.StatusInternalServerError, err
		}
	default:
		file = rawBuffer.Bytes()
	}

	// Overwrite the request Body
	r.Body = ioutil.NopCloser(bytes.NewReader(file))
	return code, nil
}

// parseFrontMatterOnlyFile parses a frontmatter only file
func parseFrontMatterOnlyFile(data interface{}, filename string) ([]byte, int, error) {
	frontmatter := strings.TrimPrefix(filepath.Ext(filename), ".")
	f, code, err := parseFrontMatter(data, frontmatter)
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

// parseFrontMatter is the frontmatter parser
func parseFrontMatter(data interface{}, frontmatter string) ([]byte, int, error) {
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

// parseCompleteFile parses a complete file
func parseCompleteFile(data map[string]interface{}, filename string, frontmatter string) ([]byte, int, error) {
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

	front, code, err := parseFrontMatter(data, frontmatter)

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

package staticgen

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	fb "github.com/filebrowser/filebrowser/lib"
	"github.com/hacdias/varutils"
)

var (
	errUnsupportedFileType = errors.New("The type of the provided file isn't supported for this action")
)

// Hugo is the Hugo static website generator.
type Hugo struct {
	// Website root
	Root string `name:"Website Root"`
	// Public folder
	Public string `name:"Public Directory"`
	// Hugo executable path
	Exe string `name:"Hugo Executable"`
	// Hugo arguments
	Args []string `name:"Hugo Arguments"`
	// Indicates if we should clean public before a new publish.
	CleanPublic bool `name:"Clean Public"`
	// previewPath is the temporary path for a preview
	previewPath string
}

// SettingsPath retrieves the correct settings path.
func (h Hugo) SettingsPath() string {
	var frontmatter string
	var err error

	if _, err = os.Stat(filepath.Join(h.Root, "config.yaml")); err == nil {
		frontmatter = "yaml"
	}

	if _, err = os.Stat(filepath.Join(h.Root, "config.json")); err == nil {
		frontmatter = "json"
	}

	if _, err = os.Stat(filepath.Join(h.Root, "config.toml")); err == nil {
		frontmatter = "toml"
	}

	if frontmatter == "" {
		return "/settings"
	}

	return "/config." + frontmatter
}

// Name is the plugin's name.
func (h Hugo) Name() string {
	return "hugo"
}

// Hook is the pre-api handler.
func (h Hugo) Hook(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	// If we are not using HTTP Post, we shall return Method Not Allowed
	// since we are only working with this method.
	if r.Method != http.MethodPost {
		return 0, nil
	}

	if c.Router != "resource" {
		return 0, nil
	}

	// We only care about creating new files from archetypes here. So...
	if r.Header.Get("Archetype") == "" {
		return 0, nil
	}

	if !c.User.AllowNew {
		return http.StatusForbidden, nil
	}

	filename := filepath.Clean(r.URL.Path)
	filename = strings.TrimPrefix(filename, string(filepath.Separator))
	archetype := r.Header.Get("archetype")

	ext := filepath.Ext(filename)

	// If the request isn't for a markdown file, we can't
	// handle it.
	if ext != ".markdown" && ext != ".md" {
		return http.StatusBadRequest, errUnsupportedFileType
	}

	// Tries to create a new file based on this archetype.
	args := []string{"new", filename, "--kind", archetype}
	if err := runCommand(h.Exe, args, h.Root); err != nil {
		return http.StatusInternalServerError, err
	}

	// Writes the location of the new file to the Header.
	w.Header().Set("Location", "/files/content/"+filename)
	return http.StatusCreated, nil
}

// Publish publishes a post.
func (h Hugo) Publish(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	filename := filepath.Join(c.User.Scope, r.URL.Path)

	// We only run undraft command if it is a file.
	if strings.HasSuffix(filename, ".md") && strings.HasSuffix(filename, ".markdown") {
		if err := h.undraft(filename); err != nil {
			return http.StatusInternalServerError, err
		}
	}

	// Regenerates the file
	h.run(false)

	return 0, nil
}

// Preview handles the preview path.
func (h *Hugo) Preview(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	// Get a new temporary path if there is none.
	if h.previewPath == "" {
		path, err := ioutil.TempDir("", "")
		if err != nil {
			return http.StatusInternalServerError, err
		}

		h.previewPath = path
	}

	// Build the arguments to execute Hugo: change the base URL,
	// build the drafts and update the destination.
	args := h.Args
	args = append(args, "--baseURL", c.RootURL()+"/preview/")
	args = append(args, "--buildDrafts")
	args = append(args, "--destination", h.previewPath)

	// Builds the preview.
	if err := runCommand(h.Exe, args, h.Root); err != nil {
		return http.StatusInternalServerError, err
	}

	// Serves the temporary path with the preview.
	http.FileServer(http.Dir(h.previewPath)).ServeHTTP(w, r)
	return 0, nil
}

func (h Hugo) run(force bool) {
	// If the CleanPublic option is enabled, clean it.
	if h.CleanPublic {
		os.RemoveAll(h.Public)
	}

	// Prevent running if watching is enabled
	if b, pos := varutils.StringInSlice("--watch", h.Args); b && !force {
		if len(h.Args) > pos && h.Args[pos+1] != "false" {
			return
		}

		if len(h.Args) == pos+1 {
			return
		}
	}

	if err := runCommand(h.Exe, h.Args, h.Root); err != nil {
		log.Println(err)
	}
}

func (h Hugo) undraft(file string) error {
	args := []string{"undraft", file}
	if err := runCommand(h.Exe, args, h.Root); err != nil && !strings.Contains(err.Error(), "not a Draft") {
		return err
	}

	return nil
}

// Setup sets up the plugin.
func (h *Hugo) Setup() error {
	var err error
	h.Exe, err = exec.LookPath("hugo")
	return err
}

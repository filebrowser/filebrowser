package filemanager

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hacdias/varutils"
	"github.com/robfig/cron"
)

var (
	// ErrUnsupportedFileType ...
	ErrUnsupportedFileType = errors.New("The type of the provided file isn't supported for this action")
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

// Hook is the pre-api handler.
func (h Hugo) Hook(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
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

	filename := filepath.Join(string(c.User.FileSystem), r.URL.Path)
	archetype := r.Header.Get("archetype")

	ext := filepath.Ext(filename)

	// If the request isn't for a markdown file, we can't
	// handle it.
	if ext != ".markdown" && ext != ".md" {
		return http.StatusBadRequest, ErrUnsupportedFileType
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
func (h Hugo) Publish(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	filename := filepath.Join(string(c.User.FileSystem), r.URL.Path)

	// Before save command handler.
	if err := c.Runner("before_publish", filename); err != nil {
		return http.StatusInternalServerError, err
	}

	// We only run undraft command if it is a file.
	if strings.HasSuffix(filename, ".md") && strings.HasSuffix(filename, ".markdown") {
		if err := h.undraft(filename); err != nil {
			return http.StatusInternalServerError, err
		}
	}

	// Regenerates the file
	h.run(false)

	// Executed the before publish command.
	if err := c.Runner("before_publish", filename); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

// Schedule schedules a post.
func (h Hugo) Schedule(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	t, err := time.Parse("2006-01-02T15:04", r.Header.Get("Schedule"))
	path := filepath.Join(string(c.User.FileSystem), r.URL.Path)
	path = filepath.Clean(path)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	scheduler := cron.New()
	scheduler.AddFunc(t.Format("05 04 15 02 01 *"), func() {
		if err := h.undraft(path); err != nil {
			log.Printf(err.Error())
		}

		h.run(false)
	})

	scheduler.Start()
	return http.StatusOK, nil
}

// Preview handles the preview path.
func (h *Hugo) Preview(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
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

func (h *Hugo) find() error {
	var err error
	if h.Exe, err = exec.LookPath("hugo"); err != nil {
		return err
	}

	return nil
}

// Jekyll is the Jekyll static website generator.
type Jekyll struct {
	// Website root
	Root string `name:"Website Root"`
	// Public folder
	Public string `name:"Public Directory"`
	// Jekyll executable path
	Exe string `name:"Executable"`
	// Jekyll arguments
	Args []string `name:"Arguments"`
	// Indicates if we should clean public before a new publish.
	CleanPublic bool `name:"Clean Public"`
	// previewPath is the temporary path for a preview
	previewPath string
}

// SettingsPath retrieves the correct settings path.
func (j Jekyll) SettingsPath() string {
	return "/_config.yml"
}

// Hook is the pre-api handler.
func (j Jekyll) Hook(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	return 0, nil
}

// Publish publishes a post.
func (j Jekyll) Publish(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	filename := filepath.Join(string(c.User.FileSystem), r.URL.Path)

	// Before save command handler.
	if err := c.Runner("before_publish", filename); err != nil {
		return http.StatusInternalServerError, err
	}

	// We only run undraft command if it is a file.
	if err := j.undraft(filename); err != nil {
		return http.StatusInternalServerError, err
	}

	// Regenerates the file
	j.run()

	// Executed the before publish command.
	if err := c.Runner("before_publish", filename); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

// Schedule schedules a post.
func (j Jekyll) Schedule(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	t, err := time.Parse("2006-01-02T15:04", r.Header.Get("Schedule"))
	path := filepath.Join(string(c.User.FileSystem), r.URL.Path)
	path = filepath.Clean(path)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	scheduler := cron.New()
	scheduler.AddFunc(t.Format("05 04 15 02 01 *"), func() {
		if err := j.undraft(path); err != nil {
			log.Printf(err.Error())
		}

		j.run()
	})

	scheduler.Start()
	return http.StatusOK, nil
}

// Preview handles the preview path.
func (j *Jekyll) Preview(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// Get a new temporary path if there is none.
	if j.previewPath == "" {
		path, err := ioutil.TempDir("", "")
		if err != nil {
			return http.StatusInternalServerError, err
		}

		j.previewPath = path
	}

	// Build the arguments to execute Hugo: change the base URL,
	// build the drafts and update the destination.
	args := j.Args
	args = append(args, "--baseurl", c.RootURL()+"/preview/")
	args = append(args, "--drafts")
	args = append(args, "--destination", j.previewPath)

	// Builds the preview.
	if err := runCommand(j.Exe, args, j.Root); err != nil {
		return http.StatusInternalServerError, err
	}

	// Serves the temporary path with the preview.
	http.FileServer(http.Dir(j.previewPath)).ServeHTTP(w, r)
	return 0, nil
}

func (j Jekyll) run() {
	// If the CleanPublic option is enabled, clean it.
	if j.CleanPublic {
		os.RemoveAll(j.Public)
	}

	if err := runCommand(j.Exe, j.Args, j.Root); err != nil {
		log.Println(err)
	}
}

func (j Jekyll) undraft(file string) error {
	if !strings.Contains(file, "_drafts") {
		return nil
	}

	return os.Rename(file, strings.Replace(file, "_drafts", "_posts", 1))
}

func (j *Jekyll) find() error {
	var err error
	if j.Exe, err = exec.LookPath("jekyll"); err != nil {
		return err
	}

	return nil
}

// runCommand executes an external command
func runCommand(command string, args []string, path string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = path
	out, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New(string(out))
	}

	return nil
}

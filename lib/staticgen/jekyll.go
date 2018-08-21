package staticgen

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	fb "github.com/filebrowser/filebrowser/lib"
)

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

// Name is the plugin's name.
func (j Jekyll) Name() string {
	return "jekyll"
}

// SettingsPath retrieves the correct settings path.
func (j Jekyll) SettingsPath() string {
	return "/_config.yml"
}

// Hook is the pre-api handler.
func (j Jekyll) Hook(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	return 0, nil
}

// Publish publishes a post.
func (j Jekyll) Publish(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	filename := filepath.Join(c.User.Scope, r.URL.Path)

	// We only run undraft command if it is a file.
	if err := j.undraft(filename); err != nil {
		return http.StatusInternalServerError, err
	}

	// Regenerates the file
	j.run()

	return 0, nil
}

// Preview handles the preview path.
func (j *Jekyll) Preview(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
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

// Setup sets up the plugin.
func (j *Jekyll) Setup() error {
	var err error
	if j.Exe, err = exec.LookPath("jekyll"); err != nil {
		return err
	}

	if len(j.Args) == 0 {
		j.Args = []string{"build"}
	}

	if j.Args[0] != "build" {
		j.Args = append([]string{"build"}, j.Args...)
	}

	return nil
}

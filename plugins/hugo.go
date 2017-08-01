package plugins

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hacdias/filemanager"
	"github.com/hacdias/varutils"
	"github.com/robfig/cron"
)

func init() {
	filemanager.RegisterPlugin("hugo", filemanager.Plugin{
		JavaScript:    hugoJavaScript,
		CommandEvents: []string{"before_publish", "after_publish"},
		Permissions: []filemanager.Permission{
			{
				Name:  "allowPublish",
				Value: true,
			},
		},
		Handler: &hugo{},
	})
}

var (
	ErrHugoNotFound        = errors.New("It seems that tou don't have 'hugo' on your PATH")
	ErrUnsupportedFileType = errors.New("The type of the provided file isn't supported for this action")
)

// Hugo is a hugo (https://gohugo.io) plugin.
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
}

// Find finds the hugo executable in the path.
func (h *Hugo) Find() error {
	var err error
	if h.Exe, err = exec.LookPath("hugo"); err != nil {
		return ErrHugoNotFound
	}

	return nil
}

// run runs Hugo with the define arguments.
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

	if err := Run(h.Exe, h.Args, h.Root); err != nil {
		log.Println(err)
	}
}

// schedule schedules a post to be published later.
func (h Hugo) schedule(c *filemanager.RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
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
			return
		}

		h.run(false)
	})

	scheduler.Start()
	return http.StatusOK, nil
}

func (h Hugo) undraft(file string) error {
	args := []string{"undraft", file}
	if err := Run(h.Exe, args, h.Root); err != nil && !strings.Contains(err.Error(), "not a Draft") {
		return err
	}

	return nil
}

type hugo struct{}

func (h hugo) Before(c *filemanager.RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	o := c.Plugins["hugo"].(*Hugo)

	// If we are using the 'magic url' for the settings, we should redirect the
	// request for the acutual path.
	if r.URL.Path == "/settings/" || r.URL.Path == "/settings" {
		var frontmatter string
		var err error

		if _, err = os.Stat(filepath.Join(o.Root, "config.yaml")); err == nil {
			frontmatter = "yaml"
		}

		if _, err = os.Stat(filepath.Join(o.Root, "config.json")); err == nil {
			frontmatter = "json"
		}

		if _, err = os.Stat(filepath.Join(o.Root, "config.toml")); err == nil {
			frontmatter = "toml"
		}

		r.URL.Path = "/config." + frontmatter
	}

	// From here on, we only care about 'hugo' router so we can bypass
	// the others.
	if c.Router != "hugo" {
		return 0, nil
	}

	// If we are not using HTTP Post, we shall return Method Not Allowed
	// since we are only working with this method.
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, nil
	}

	// If we are creating a file built from an archetype.
	if r.Header.Get("Archetype") != "" {
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
		if err := Run(o.Exe, args, o.Root); err != nil {
			return http.StatusInternalServerError, err
		}

		// Writes the location of the new file to the Header.
		w.Header().Set("Location", "/files/content/"+filename)
		return http.StatusCreated, nil
	}

	// If we are trying to regenerate the website.
	if r.Header.Get("Regenerate") == "true" {
		if !c.User.Permissions["allowPublish"] {
			return http.StatusForbidden, nil
		}

		filename := filepath.Join(string(c.User.FileSystem), r.URL.Path)

		// Before save command handler.
		if err := c.Runner("before_publish", filename); err != nil {
			return http.StatusInternalServerError, err
		}

		// We only run undraft command if it is a file.
		if strings.HasSuffix(filename, ".md") && strings.HasSuffix(filename, ".markdown") {
			if err := o.undraft(filename); err != nil {
				return http.StatusInternalServerError, err
			}

		}

		// Regenerates the file
		o.run(false)

		// Executed the before publish command.
		if err := c.Runner("before_publish", filename); err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}

	if r.Header.Get("Schedule") != "" {
		if !c.User.Permissions["allowPublish"] {
			return http.StatusForbidden, nil
		}

		return o.schedule(c, w, r)
	}

	return http.StatusNotFound, nil
}

func (h hugo) After(c *filemanager.RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	return 0, nil
}

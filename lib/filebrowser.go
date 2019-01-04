package lib

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mholt/caddy"
	"github.com/spf13/afero"
)

var defaultEvents = []string{
	"save",
	"copy",
	"rename",
	"upload",
	"delete",
}

// FileBrowser represents a File Browser instance which must
// be created through NewFileBrowser.
type FileBrowser struct {
	settings *Settings
	storage  StorageBackend
	mux      sync.RWMutex
}

// NewFileBrowser creates a new File Browser instance from a
// storage backend. If that backend doesn't contain settings
// on it (returns ErrNotExist), then we generate a new key
// and base settings.
func NewFileBrowser(backend StorageBackend) (*FileBrowser, error) {
	settings, err := backend.GetSettings()

	if err == ErrNotExist {
		var key []byte
		key, err = generateRandomBytes(64)

		if err != nil {
			return nil, err
		}

		settings = &Settings{Key: key}
		err = backend.SaveSettings(settings)
	}

	if err != nil {
		return nil, err
	}

	return &FileBrowser{
		settings: settings,
		storage:  backend,
	}, nil
}

// RLockSettings locks the settings for reading.
func (f *FileBrowser) RLockSettings() {
	f.mux.RLock()
}

// RUnlockSettings unlocks the settings for reading.
func (f *FileBrowser) RUnlockSettings() {
	f.mux.RUnlock()
}

// RulesCheck matches a path against the user rules and the
// global rules. Returns true if allowed, false if not.
func (f *FileBrowser) RulesCheck(u *User, path string) bool {
	for _, rule := range u.Rules {
		if rule.Matches(path) {
			return rule.Allow
		}
	}

	f.mux.RLock()
	defer f.mux.RUnlock()

	for _, rule := range f.settings.Rules {
		if rule.Matches(path) {
			return rule.Allow
		}
	}

	return true
}

// RunHook runs the hooks for the before and after event.
func (f *FileBrowser) RunHook(fn func() error, evt, path, dst string, user *User) error {
	path = user.FullPath(path)
	dst = user.FullPath(dst)

	if val, ok := f.settings.Commands["before_"+evt]; ok {
		for _, command := range val {
			err := f.exec(command, "before_"+evt, path, dst, user)
			if err != nil {
				return err
			}
		}
	}

	err := fn()
	if err != nil {
		return err
	}

	if val, ok := f.settings.Commands["after_"+evt]; ok {
		for _, command := range val {
			err := f.exec(command, "after_"+evt, path, dst, user)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ParseCommand parses the command taking in account if the current
// instance uses a shell to run the commands or just calls the binary
// directyly.
func (f *FileBrowser) ParseCommand(raw string) ([]string, error) {
	f.RLockSettings()
	defer f.RUnlockSettings()

	command := []string{}

	if len(f.settings.Shell) == 0 {
		cmd, args, err := caddy.SplitCommandAndArgs(raw)
		if err != nil {
			return nil, err
		}

		_, err = exec.LookPath(cmd)
		if err != nil {
			return nil, err
		}

		command = append(command, cmd)
		command = append(command, args...)
	} else {
		command = append(f.settings.Shell, raw)
	}

	return command, nil
}

// ApplyDefaults applies the default options to a user.
func (f *FileBrowser) ApplyDefaults(u *User) {
	f.RLockSettings()
	u.Scope = f.settings.Defaults.Scope
	u.Locale = f.settings.Defaults.Locale
	u.ViewMode = f.settings.Defaults.ViewMode
	u.Perm = f.settings.Defaults.Perm
	u.Sorting = f.settings.Defaults.Sorting
	u.Commands = f.settings.Defaults.Commands
	f.RUnlockSettings()
}

// NewFile creates a File object from a path and a given user. This File
// object will be automatically filled depending on if it is a directory
// or a file. If it's a video file, it will also detect any subtitles.
func (f *FileBrowser) NewFile(path string, user *User) (*File, error) {
	if !f.RulesCheck(user, path) {
		return nil, os.ErrPermission
	}

	info, err := user.Fs.Stat(path)
	if err != nil {
		return nil, err
	}

	file := &File{
		Path:      path,
		Name:      info.Name(),
		ModTime:   info.ModTime(),
		Mode:      info.Mode(),
		IsDir:     info.IsDir(),
		Size:      info.Size(),
		Extension: filepath.Ext(info.Name()),
	}

	if file.IsDir {
		return file, f.readListing(file, user)
	}

	err = f.detectType(file, user)
	if err != nil {
		return nil, err
	}

	if file.Type == "video" {
		f.detectSubtitles(file, user)
	}

	return file, err
}

// Checksum checksums a given File for a given User, using a specific
// algorithm. The checksums data is saved on File object.
func (f *FileBrowser) Checksum(file *File, user *User, algo string) error {
	if file.IsDir {
		return ErrIsDirectory
	}

	if file.Checksums == nil {
		file.Checksums = map[string]string{}
	}

	i, err := user.Fs.Open(file.Path)
	if err != nil {
		return err
	}
	defer i.Close()

	var h hash.Hash

	switch algo {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return ErrInvalidOption
	}

	_, err = io.Copy(h, i)
	if err != nil {
		return err
	}

	file.Checksums[algo] = hex.EncodeToString(h.Sum(nil))
	return nil
}

func (f *FileBrowser) readListing(file *File, user *User) error {
	afs := &afero.Afero{Fs: user.Fs}
	files, err := afs.ReadDir(file.Path)
	if err != nil {
		return err
	}

	listing := &Listing{
		Items:    []*File{},
		NumDirs:  0,
		NumFiles: 0,
	}

	for _, i := range files {
		name := i.Name()
		path := path.Join(file.Path, name)

		if !f.RulesCheck(user, path) {
			continue
		}

		if strings.HasPrefix(i.Mode().String(), "L") {
			// It's a symbolic link. We try to follow it. If it doesn't work,
			// we stay with the link information instead if the target's.
			info, err := os.Stat(name)
			if err == nil {
				i = info
			}
		}

		file := &File{
			Name:      name,
			Size:      i.Size(),
			ModTime:   i.ModTime(),
			Mode:      i.Mode(),
			IsDir:     i.IsDir(),
			Extension: filepath.Ext(name),
			Path:      path,
		}

		if file.IsDir {
			listing.NumDirs++
		} else {
			listing.NumFiles++

			err := f.detectType(file, user)
			if err != nil {
				return err
			}
		}

		listing.Items = append(listing.Items, file)
	}

	file.Listing = listing
	return nil
}

func (f *FileBrowser) detectType(file *File, user *User) error {
	i, err := user.Fs.Open(file.Path)
	if err != nil {
		return err
	}
	defer i.Close()

	buffer := make([]byte, 512)
	n, err := i.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	mimetype := mime.TypeByExtension(file.Extension)
	if mimetype == "" {
		mimetype = http.DetectContentType(buffer[:n])
	}

	switch {
	case strings.HasPrefix(mimetype, "video"):
		file.Type = "video"
		return nil
	case strings.HasPrefix(mimetype, "audio"):
		file.Type = "audio"
		return nil
	case strings.HasPrefix(mimetype, "image"):
		file.Type = "image"
		return nil
	case isBinary(string(buffer[:n])) || file.Size > 10*1024*1024: // 10 MB
		file.Type = "blob"
		return nil
	default:
		file.Type = "text"
		afs := &afero.Afero{Fs: user.Fs}
		content, err := afs.ReadFile(file.Path)
		if err != nil {
			return err
		}
		file.Content = string(content)
	}

	if !user.Perm.Modify && file.Type == "text" {
		file.Type = "textImmutable"
	}

	return nil
}

func (f *FileBrowser) detectSubtitles(file *File, user *User) {
	file.Subtitles = []string{}
	ext := filepath.Ext(file.Path)
	base := strings.TrimSuffix(file.Path, ext)

	// TODO: detect multiple languages. Like base.lang.vtt

	path := base + ".vtt"
	if _, err := user.Fs.Stat(path); err == nil {
		file.Subtitles = append(file.Subtitles, path)
	}
}

func (f *FileBrowser) exec(raw, evt, path, dst string, user *User) error {
	blocking := true

	if strings.HasSuffix(raw, "&") {
		blocking = false
		raw = strings.TrimSpace(strings.TrimSuffix(raw, "&"))
	}

	command, err := f.ParseCommand(raw)
	if err != nil {
		return err
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("FILE=%s", path))
	cmd.Env = append(cmd.Env, fmt.Sprintf("SCOPE=%s", user.Scope))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TRIGGER=%s", evt))
	cmd.Env = append(cmd.Env, fmt.Sprintf("USERNAME=%s", user.Username))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DESTINATION=%s", dst))

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if !blocking {
		log.Printf("[INFO] Nonblocking Command: \"%s\"", strings.Join(command, " "))
		return cmd.Start()
	}

	log.Printf("[INFO] Blocking Command: \"%s\"", strings.Join(command, " "))
	return cmd.Run()
}

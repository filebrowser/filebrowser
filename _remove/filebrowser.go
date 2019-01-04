package lib

/*
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

	"github.com/filebrowser/filebrowser/files"
	"github.com/filebrowser/filebrowser/rules"
	"github.com/filebrowser/filebrowser/settings"
	"github.com/filebrowser/filebrowser/users"
	"github.com/mholt/caddy"
	"github.com/spf13/afero"
)



// FileBrowser represents a File Browser instance which must
// be created through NewFileBrowser.
type FileBrowser struct {
	settings *settings.Settings
	mux      sync.RWMutex
}

// NewFileBrowser creates a new File Browser instance from a
// storage backend. If that backend doesn't contain settings
// on it (returns ErrNotExist), then we generate a new key
// and base settings.
func NewFileBrowser(backend StorageBackend) (*FileBrowser, error) {
	set, err := backend.GetSettings()

	if err == ErrNotExist {
		var key []byte
		key, err = generateRandomBytes(64)

		if err != nil {
			return nil, err
		}

		set = &settings.Settings{Key: key}
		err = backend.SaveSettings(set)
	}

	if err != nil {
		return nil, err
	}

	return &FileBrowser{
		settings: set,
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

// CheckRules matches the rules against user rules and global rules.
func (f *FileBrowser) CheckRules(path string, user *users.User) bool {
	f.RLockSettings()
	val := rules.Check(path, user, f.settings)
	f.RUnlockSettings()
	return val
}

// RunHook runs the hooks for the before and after event.
func (f *FileBrowser) RunHook(fn func() error, evt, path, dst string, user *users.User) error {
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







func (f *FileBrowser) exec(raw, evt, path, dst string, user *users.User) error {
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
*/

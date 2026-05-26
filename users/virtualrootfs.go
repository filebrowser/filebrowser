package users

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// VirtualRootFs presents multiple scopes as folders under a virtual root.
// Each scope appears as a top-level directory named by its basename.
// e.g. scopes ["/data/projects", "/data/media"] become:
//
//	/ (virtual root)
//	├── projects/  → BasePathFs("/data/projects")
//	└── media/     → BasePathFs("/data/media")
type VirtualRootFs struct {
	scopes map[string]afero.Fs // basename → scoped filesystem
	order  []string            // ordered basenames for consistent listing
}

// NewVirtualRootFs creates a virtual root filesystem from multiple scopes.
// baseScope is the server root, scopes are the user's scope paths.
func NewVirtualRootFs(baseScope string, scopes []string) *VirtualRootFs {
	scopeMap := make(map[string]afero.Fs, len(scopes))
	order := make([]string, 0, len(scopes))

	for _, scope := range scopes {
		absPath := filepath.Join(baseScope, filepath.Join("/", scope))
		baseName := ScopeBaseName(scope)
		scopeMap[baseName] = afero.NewBasePathFs(afero.NewOsFs(), absPath)
		order = append(order, baseName)
	}

	return &VirtualRootFs{
		scopes: scopeMap,
		order:  order,
	}
}

// resolve splits a path into scope name and remaining path within that scope.
// Returns the scope's filesystem and the relative path within it.
func (v *VirtualRootFs) resolve(name string) (afero.Fs, string, error) {
	name = path.Clean("/" + name)
	if name == "/" {
		return nil, "/", nil
	}

	// Strip leading slash and split
	trimmed := strings.TrimPrefix(name, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	scopeName := parts[0]

	scopeFs, ok := v.scopes[scopeName]
	if !ok {
		return nil, "", &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}

	remaining := "/"
	if len(parts) > 1 {
		remaining = "/" + parts[1]
	}

	return scopeFs, remaining, nil
}

// virtualRootFileInfo represents the virtual root directory.
type virtualRootFileInfo struct{}

func (virtualRootFileInfo) Name() string      { return "/" }
func (virtualRootFileInfo) Size() int64        { return 0 }
func (virtualRootFileInfo) Mode() fs.FileMode  { return os.ModeDir | 0o755 }
func (virtualRootFileInfo) ModTime() time.Time { return time.Time{} }
func (virtualRootFileInfo) IsDir() bool        { return true }
func (virtualRootFileInfo) Sys() interface{}   { return nil }

// virtualScopeFileInfo represents a scope directory entry at the root level.
type virtualScopeFileInfo struct {
	name string
	info os.FileInfo
}

func (f *virtualScopeFileInfo) Name() string      { return f.name }
func (f *virtualScopeFileInfo) Size() int64        { return f.info.Size() }
func (f *virtualScopeFileInfo) Mode() fs.FileMode  { return f.info.Mode() }
func (f *virtualScopeFileInfo) ModTime() time.Time { return f.info.ModTime() }
func (f *virtualScopeFileInfo) IsDir() bool        { return true }
func (f *virtualScopeFileInfo) Sys() interface{}   { return f.info.Sys() }

// virtualRootDir implements afero.File for the virtual root directory listing.
type virtualRootDir struct {
	entries []os.FileInfo
	pos     int
}

func (d *virtualRootDir) Close() error                             { return nil }
func (d *virtualRootDir) Read(_ []byte) (int, error)               { return 0, os.ErrInvalid }
func (d *virtualRootDir) ReadAt(_ []byte, _ int64) (int, error)    { return 0, os.ErrInvalid }
func (d *virtualRootDir) Seek(_ int64, _ int) (int64, error)       { return 0, os.ErrInvalid }
func (d *virtualRootDir) Write(_ []byte) (int, error)              { return 0, os.ErrInvalid }
func (d *virtualRootDir) WriteAt(_ []byte, _ int64) (int, error)   { return 0, os.ErrInvalid }
func (d *virtualRootDir) Name() string                             { return "/" }
func (d *virtualRootDir) Stat() (os.FileInfo, error)               { return virtualRootFileInfo{}, nil }
func (d *virtualRootDir) Sync() error                              { return nil }
func (d *virtualRootDir) Truncate(_ int64) error                   { return os.ErrInvalid }
func (d *virtualRootDir) WriteString(_ string) (int, error)        { return 0, os.ErrInvalid }
func (d *virtualRootDir) Readdir(count int) ([]os.FileInfo, error) { return d.readdirImpl(count) }
func (d *virtualRootDir) Readdirnames(count int) ([]string, error) {
	infos, err := d.readdirImpl(count)
	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name()
	}
	return names, err
}

func (d *virtualRootDir) readdirImpl(count int) ([]os.FileInfo, error) {
	if count <= 0 {
		result := d.entries[d.pos:]
		d.pos = len(d.entries)
		return result, nil
	}

	end := d.pos + count
	if end > len(d.entries) {
		end = len(d.entries)
	}
	result := d.entries[d.pos:end]
	d.pos = end
	return result, nil
}

// -- afero.Fs interface implementation --

func (v *VirtualRootFs) Name() string { return "VirtualRootFs" }

func (v *VirtualRootFs) Create(name string) (afero.File, error) {
	scopeFs, relPath, err := v.resolve(name)
	if err != nil {
		return nil, err
	}
	if scopeFs == nil {
		return nil, os.ErrPermission // can't create in virtual root
	}
	return scopeFs.Create(relPath)
}

func (v *VirtualRootFs) Mkdir(name string, perm os.FileMode) error {
	scopeFs, relPath, err := v.resolve(name)
	if err != nil {
		return err
	}
	if scopeFs == nil {
		return os.ErrPermission
	}
	return scopeFs.Mkdir(relPath, perm)
}

func (v *VirtualRootFs) MkdirAll(name string, perm os.FileMode) error {
	scopeFs, relPath, err := v.resolve(name)
	if err != nil {
		return err
	}
	if scopeFs == nil {
		return os.ErrPermission
	}
	return scopeFs.MkdirAll(relPath, perm)
}

func (v *VirtualRootFs) Open(name string) (afero.File, error) {
	scopeFs, relPath, err := v.resolve(name)
	if err != nil {
		return nil, err
	}
	if scopeFs == nil {
		// Opening the virtual root: return directory listing
		return v.openRoot()
	}
	return scopeFs.Open(relPath)
}

func (v *VirtualRootFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	scopeFs, relPath, err := v.resolve(name)
	if err != nil {
		return nil, err
	}
	if scopeFs == nil {
		if flag == os.O_RDONLY {
			return v.openRoot()
		}
		return nil, os.ErrPermission
	}
	return scopeFs.OpenFile(relPath, flag, perm)
}

func (v *VirtualRootFs) Remove(name string) error {
	scopeFs, relPath, err := v.resolve(name)
	if err != nil {
		return err
	}
	if scopeFs == nil || relPath == "/" {
		return os.ErrPermission
	}
	return scopeFs.Remove(relPath)
}

func (v *VirtualRootFs) RemoveAll(name string) error {
	scopeFs, relPath, err := v.resolve(name)
	if err != nil {
		return err
	}
	if scopeFs == nil || relPath == "/" {
		return os.ErrPermission
	}
	return scopeFs.RemoveAll(relPath)
}

func (v *VirtualRootFs) Rename(oldname, newname string) error {
	oldFs, oldRel, err := v.resolve(oldname)
	if err != nil {
		return err
	}
	newFs, newRel, err := v.resolve(newname)
	if err != nil {
		return err
	}
	if oldFs == nil || newFs == nil {
		return os.ErrPermission
	}
	if oldRel == "/" || newRel == "/" {
		return os.ErrPermission
	}
	// Can only rename within the same scope
	if oldFs != newFs {
		return &os.PathError{Op: "rename", Path: oldname, Err: os.ErrPermission}
	}
	return oldFs.Rename(oldRel, newRel)
}

func (v *VirtualRootFs) Stat(name string) (os.FileInfo, error) {
	scopeFs, relPath, err := v.resolve(name)
	if err != nil {
		return nil, err
	}
	if scopeFs == nil {
		return virtualRootFileInfo{}, nil
	}
	if relPath == "/" {
		// Stat on a scope root — return info with the scope's display name
		info, statErr := scopeFs.Stat("/")
		if statErr != nil {
			return nil, statErr
		}
		scopeName := strings.TrimPrefix(path.Clean("/"+name), "/")
		scopeName = strings.SplitN(scopeName, "/", 2)[0]
		return &virtualScopeFileInfo{name: scopeName, info: info}, nil
	}
	return scopeFs.Stat(relPath)
}

func (v *VirtualRootFs) Chmod(name string, mode os.FileMode) error {
	scopeFs, relPath, err := v.resolve(name)
	if err != nil {
		return err
	}
	if scopeFs == nil {
		return os.ErrPermission
	}
	return scopeFs.Chmod(relPath, mode)
}

func (v *VirtualRootFs) Chown(name string, uid, gid int) error {
	scopeFs, relPath, err := v.resolve(name)
	if err != nil {
		return err
	}
	if scopeFs == nil {
		return os.ErrPermission
	}
	return scopeFs.Chown(relPath, uid, gid)
}

func (v *VirtualRootFs) Chtimes(name string, atime, mtime time.Time) error {
	scopeFs, relPath, err := v.resolve(name)
	if err != nil {
		return err
	}
	if scopeFs == nil {
		return os.ErrPermission
	}
	return scopeFs.Chtimes(relPath, atime, mtime)
}

func (v *VirtualRootFs) openRoot() (afero.File, error) {
	entries := make([]os.FileInfo, 0, len(v.order))
	for _, name := range v.order {
		scopeFs := v.scopes[name]
		info, err := scopeFs.Stat("/")
		if err != nil {
			continue
		}
		entries = append(entries, &virtualScopeFileInfo{name: name, info: info})
	}
	return &virtualRootDir{entries: entries}, nil
}

// RealPath resolves a virtual path to the actual OS filesystem path.
// This is needed by FileInfo.RealPath() and User.FullPath().
func (v *VirtualRootFs) RealPath(name string) (string, error) {
	scopeFs, relPath, err := v.resolve(name)
	if err != nil {
		return "", err
	}
	if scopeFs == nil {
		return "", os.ErrPermission
	}

	// Delegate to the underlying BasePathFs
	if bp, ok := scopeFs.(*afero.BasePathFs); ok {
		return bp.RealPath(relPath)
	}

	return "", os.ErrPermission
}

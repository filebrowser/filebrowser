package types

import (
	"os"
	"path"
	"syscall"
	"time"

	"github.com/spf13/afero"
)

type userFs struct {
	source   afero.Fs
	user     *User
	settings *Settings
}

type userFile struct {
	f    afero.File
	path string
	fs   *userFs
}

func (u *userFs) isAllowed(name string) bool {
	if !isAllowed(name, u.user.Rules) {
		return false
	}

	return isAllowed(name, u.settings.Rules)
}

func (u *userFs) FullPath(path string) string {
	return afero.FullBaseFsPath(u.source.(*afero.BasePathFs), path)
}

func (u *userFs) Chtimes(name string, a, m time.Time) error {
	if !u.isAllowed(name) {
		return syscall.ENOENT
	}

	return u.source.Chtimes(name, a, m)
}

func (u *userFs) Chmod(name string, mode os.FileMode) error {
	if !u.isAllowed(name) {
		return syscall.ENOENT
	}

	return u.source.Chmod(name, mode)
}

func (u *userFs) Name() string {
	return "userFs"
}

func (u *userFs) Stat(name string) (os.FileInfo, error) {
	if !u.isAllowed(name) {
		return nil, syscall.ENOENT
	}

	return u.source.Stat(name)
}

func (u *userFs) Rename(oldname, newname string) error {
	if !u.user.Perm.Rename {
		return os.ErrPermission
	}

	if !u.isAllowed(oldname) || !u.isAllowed(newname) {
		return syscall.ENOENT
	}

	return u.source.Rename(oldname, newname)
}

func (u *userFs) RemoveAll(name string) error {
	if !u.user.Perm.Delete {
		return os.ErrPermission
	}

	if !u.isAllowed(name) {
		return syscall.ENOENT
	}

	return u.source.RemoveAll(name)
}

func (u *userFs) Remove(name string) error {
	if !u.user.Perm.Delete {
		return os.ErrPermission
	}

	if !u.isAllowed(name) {
		return syscall.ENOENT
	}

	return u.source.Remove(name)
}

func (u *userFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if !u.isAllowed(name) {
		return nil, syscall.ENOENT
	}

	return u.source.OpenFile(name, flag, perm)
}

func (u *userFs) Open(name string) (afero.File, error) {
	if !u.isAllowed(name) {
		return nil, syscall.ENOENT
	}

	f, err := u.source.Open(name)
	return &userFile{fs: u, path: name, f: f}, err
}

func (u *userFs) Mkdir(name string, perm os.FileMode) error {
	if !u.user.Perm.Create {
		return os.ErrPermission
	}

	if !u.isAllowed(name) {
		return syscall.ENOENT
	}

	return u.source.Mkdir(name, perm)
}

func (u *userFs) MkdirAll(name string, perm os.FileMode) error {
	if !u.user.Perm.Create {
		return os.ErrPermission
	}

	if !u.isAllowed(name) {
		return syscall.ENOENT
	}

	return u.source.MkdirAll(name, perm)
}

func (u *userFs) Create(name string) (afero.File, error) {
	if !u.user.Perm.Create {
		return nil, os.ErrPermission
	}

	if !u.isAllowed(name) {
		return nil, syscall.ENOENT
	}

	return u.source.Create(name)
}

func (f *userFile) Close() error {
	return f.f.Close()
}

func (f *userFile) Read(s []byte) (int, error) {
	return f.f.Read(s)
}

func (f *userFile) ReadAt(s []byte, o int64) (int, error) {
	return f.f.ReadAt(s, o)
}

func (f *userFile) Seek(o int64, w int) (int64, error) {
	return f.f.Seek(o, w)
}

func (f *userFile) Write(s []byte) (int, error) {
	return f.f.Write(s)
}

func (f *userFile) WriteAt(s []byte, o int64) (int, error) {
	return f.f.WriteAt(s, o)
}

func (f *userFile) Name() string {
	return f.f.Name()
}

func (f *userFile) Readdir(c int) (fi []os.FileInfo, err error) {
	var rfi []os.FileInfo
	rfi, err = f.f.Readdir(c)
	if err != nil {
		return nil, err
	}
	for _, i := range rfi {
		if f.fs.isAllowed(path.Join(f.path, i.Name())) {
			fi = append(fi, i)
		}
	}
	return fi, nil
}

func (f *userFile) Readdirnames(c int) (n []string, err error) {
	fi, err := f.Readdir(c)
	if err != nil {
		return nil, err
	}
	for _, s := range fi {
		if f.fs.isAllowed(s.Name()) {
			n = append(n, s.Name())
		}
	}
	return n, nil
}

func (f *userFile) Stat() (os.FileInfo, error) {
	return f.f.Stat()
}

func (f *userFile) Sync() error {
	return f.f.Sync()
}

func (f *userFile) Truncate(s int64) error {
	return f.f.Truncate(s)
}

func (f *userFile) WriteString(s string) (int, error) {
	return f.f.WriteString(s)
}

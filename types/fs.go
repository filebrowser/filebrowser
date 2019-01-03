package types

import (
	"os"
	"syscall"
	"time"

	"github.com/spf13/afero"
)

type userFs struct {
	source   afero.Fs
	user     *User
	settings *Settings
}

func (u *userFs) isAllowed(name string) bool {
	if !isAllowed(name, u.user.Rules) {
		return false
	}

	return isAllowed(name, u.settings.Rules)
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

	return u.source.Open(name)
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

package fbhttp

import (
	"context"
	"os"

	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

type webDavFS struct {
	fs afero.Fs
}

func (w *webDavFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return w.fs.MkdirAll(name, perm)
}

func (w *webDavFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	f, err := w.fs.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (w *webDavFS) RemoveAll(ctx context.Context, name string) error {
	return w.fs.RemoveAll(name)
}

func (w *webDavFS) Rename(ctx context.Context, oldName, newName string) error {
	return w.fs.Rename(oldName, newName)
}

func (w *webDavFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return w.fs.Stat(name)
}

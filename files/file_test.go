package files

import (
	"path"
	"testing"

	"github.com/spf13/afero"
)

type checkerFunc func(string) bool

func (f checkerFunc) Check(filePath string) bool {
	return f(filePath)
}

func TestReadListingCalculatesDirectorySizes(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	writeTestFile(t, fs, "/notes.txt", "hi")
	writeTestFile(t, fs, "/projects/api/main.go", "123")
	writeTestFile(t, fs, "/projects/web/index.html", "12345")

	file, err := NewFileInfo(&FileOptions{
		Fs:                 fs,
		Path:               "/",
		Expand:             true,
		ShowDirectorySizes: true,
		Checker:            checkerFunc(func(string) bool { return true }),
	})
	if err != nil {
		t.Fatalf("NewFileInfo returned error: %v", err)
	}

	projects := findItemByName(t, file.Items, "projects")
	if got, want := projects.Size, int64(8); got != want {
		t.Fatalf("directory size = %d, want %d", got, want)
	}

	if got, want := file.Size, int64(10); got != want {
		t.Fatalf("root directory size = %d, want %d", got, want)
	}
}

func TestReadListingSkipsHiddenEntriesInDirectorySizes(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	writeTestFile(t, fs, "/visible/report.txt", "123")
	writeTestFile(t, fs, "/visible/.secret.txt", "12345")
	writeTestFile(t, fs, "/visible/.private/hidden.txt", "1234567")

	file, err := NewFileInfo(&FileOptions{
		Fs:                 fs,
		Path:               "/",
		Expand:             true,
		ShowDirectorySizes: true,
		Checker: checkerFunc(func(filePath string) bool {
			for current := filePath; current != "/" && current != "."; current = path.Dir(current) {
				if path.Base(current) != "" && path.Base(current)[0] == '.' {
					return false
				}
			}
			return true
		}),
	})
	if err != nil {
		t.Fatalf("NewFileInfo returned error: %v", err)
	}

	visible := findItemByName(t, file.Items, "visible")
	if got, want := visible.Size, int64(3); got != want {
		t.Fatalf("directory size with hidden files filtered = %d, want %d", got, want)
	}
}

func TestApplySortUsesCalculatedDirectorySizes(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	writeTestFile(t, fs, "/small/a.txt", "1")
	writeTestFile(t, fs, "/large/a.txt", "123456")

	file, err := NewFileInfo(&FileOptions{
		Fs:                 fs,
		Path:               "/",
		Expand:             true,
		ShowDirectorySizes: true,
		Checker:            checkerFunc(func(string) bool { return true }),
	})
	if err != nil {
		t.Fatalf("NewFileInfo returned error: %v", err)
	}

	file.Sorting = Sorting{By: "size", Asc: true}
	file.ApplySort()

	if got, want := file.Items[0].Name, "small"; got != want {
		t.Fatalf("first item after sort = %q, want %q", got, want)
	}
	if got, want := file.Items[1].Name, "large"; got != want {
		t.Fatalf("second item after sort = %q, want %q", got, want)
	}
}

func writeTestFile(t *testing.T, fs afero.Fs, filePath, contents string) {
	t.Helper()

	if err := fs.MkdirAll(path.Dir(filePath), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", path.Dir(filePath), err)
	}

	if err := afero.WriteFile(fs, filePath, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) returned error: %v", filePath, err)
	}
}

func findItemByName(t *testing.T, items []*FileInfo, name string) *FileInfo {
	t.Helper()

	for _, item := range items {
		if item.Name == name {
			return item
		}
	}

	t.Fatalf("item %q not found", name)
	return nil
}
